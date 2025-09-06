package wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/blockchain"
	"github.com/sagarkarki99/internal/keychain"
	repo "github.com/sagarkarki99/internal/repository"
)

type Input struct {
	Txid         string  `json:"txid"`
	Amount       float64 `json:"amount"`
	ScriptPubKey string  `json:"scriptPubKey"`
	Vout         int     `json:"vout"`
}

func (i *Input) AmountInSatoshi() int64 {
	return int64(i.Amount * 1e8)
}

type Utxo struct {
	Inputs      []Input `json:"unspents"`
	TotalAmount float64 `json:"total_amount"`
}

type WalletService interface {
	GetDepositAddress(userId int) string
	GetBalance(addr string) float64
	SendToAddress(userId int, amount float64, destinationAddress string) error
}

func NewWalletService(kc keychain.Keychain) WalletService {
	return &WalletServiceImpl{
		repo: repo.New(),
		kc:   kc,
	}
}

type WalletServiceImpl struct {
	repo repo.WalletRepository
	kc   keychain.Keychain
}

func (ws *WalletServiceImpl) GetDepositAddress(userId int) string {
	var address *db.Address
	var account *db.Account
	account, _ = ws.repo.GetAccount(userId)
	if account == nil {
		ai, err := ws.kc.GenerateAddress(uint32(userId))
		if err != nil {
			return ""
		}

		a := &db.Account{
			XpubKey:     ai.Xpub,
			Fingerprint: ai.Fingerprint,
		}

		id, err := ws.repo.SaveAccount(*a)
		if err != nil {
			return ""
		}
		account = a
		account.Id = id

		address = &db.Address{
			Index:     0,
			NextIndex: 1,
			AccountId: account.Id,
		}

		payload, _ := ws.getDescriptorPayload(account.Fingerprint, account.XpubKey)

		_, e := blockchain.QueryFromBytes("importdescriptors", payload)
		if e != nil {
			slog.Error("Error importing address", "error", e)
		}
	}

	if address == nil {
		add, _ := ws.repo.GetAddress(account.Id)
		address = &db.Address{
			Index:     add.NextIndex,
			NextIndex: add.NextIndex + 1,
			AccountId: add.AccountId,
		}
	}

	extKey, _ := hdkeychain.NewKeyFromString(account.XpubKey)

	// We received m/84h/0h/0h from addressInfo.
	// Adding change type (0 for external, 1 for internal)
	//Path: m/84h/0h/0h/0
	changeKey, _ := extKey.Derive(0)

	// Adding index for address generation (0,1,2...)
	//Path: m/84h/0h/0h/0/0

	// we need to import m/84h/0h/0h/xpub/0/* to the node.
	childKey, err := changeKey.Derive(uint32(address.Index))
	if err != nil {
		slog.Error("Error deriving child key", "error", err)
	}

	childPub, _ := childKey.ECPubKey()

	addr, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(childPub.SerializeCompressed()), blockchain.GetNetworkParams())
	fmt.Printf("Address ( m/84h/1h/%dh/0/%d): %s", address.AccountId, address.Index, addr.EncodeAddress())
	fmt.Println("--------------------------------")
	address.Hash = addr.EncodeAddress()

	ws.repo.SaveAddress(*address)

	return address.Hash
}

func (ws *WalletServiceImpl) getDescriptorPayload(fp string, xpub string) ([]byte, error) {

	p := fmt.Sprintf("wpkh([%s/84h/1h/0h]%s/0/*)", fp, xpub)
	data, err := blockchain.Query("getdescriptorinfo", []interface{}{p})
	if err != nil {
		return nil, err
	}
	var dataMap map[string]interface{}
	r, _ := data.MarshalJSON()

	json.Unmarshal(r, &dataMap)
	des := fmt.Sprintf("%s#%s", p, dataMap["checksum"].(string))
	payload := map[string]interface{}{
		"desc":      des,
		"timestamp": "now",
		"watchonly": true,
		"range":     []int{0, 1000},
	}
	payloadBytes, err := json.Marshal([]interface{}{payload})
	if err != nil {
		slog.Error("Error marshaling payloads", "error", err.Error())
		return nil, err
	}
	return payloadBytes, err
}

func (ws *WalletServiceImpl) GetBalance(addr string) float64 {
	utxo, err := ws.getUTXOs(addr)
	if err != nil {
		fmt.Println("Error getting UTXOs: ", err)
		return 0
	}

	fmt.Println("Total Amount: ", utxo.TotalAmount)
	return utxo.TotalAmount

}

func (*WalletServiceImpl) getUTXOs(addr string) (utxo *Utxo, err error) {
	address := "addr(" + addr + ")"
	params := []interface{}{"start", []string{address}}
	res, _ := blockchain.Query("scantxoutset", params)
	r, _ := res.MarshalJSON()

	if err := json.Unmarshal(r, &utxo); err != nil {
		fmt.Println("Error unmarshaling UTXO response:", err)
		return nil, err
	}
	return utxo, nil
}

func (ws *WalletServiceImpl) SendToAddress(userId int, amount float64, destinationAddress string) error {
	sender, err := ws.repo.GetAddress(userId)
	if err != nil {
		return errors.New("user do not have any wallet")
	}
	myAddress, _ := btcutil.DecodeAddress(sender.Hash, blockchain.GetNetworkParams())
	fmt.Println("Sender Address: ", sender.Hash)
	u, err := ws.getUTXOs(myAddress.EncodeAddress())
	if err != nil {
		fmt.Println("Error getting UTXOs: ", err)
		return err
	}

	if u.TotalAmount < amount {
		return fmt.Errorf("insufficient funds")
	}

	amountInSatoshi := int64(amount * 1e8)

	tx := wire.NewMsgTx(wire.TxVersion)

	// create Input
	var usedInputs []Input
	totalAmount := int64(0)

	for _, v := range u.Inputs {
		h, _ := chainhash.NewHashFromStr(v.Txid)
		fmt.Println("ScriptPubKey: ", v.ScriptPubKey)
		if totalAmount >= amountInSatoshi {
			// Need to do change and
			break
		}
		usedInputs = append(usedInputs, v)
		totalAmount += int64(v.AmountInSatoshi())
		op := wire.NewOutPoint(h, uint32(v.Vout))
		inp := wire.NewTxIn(op, []byte{}, nil)
		tx.AddTxIn(inp)

	}

	// create Sender output
	dAddr, _ := btcutil.DecodeAddress(destinationAddress, blockchain.GetNetworkParams())
	pubScript, err := txscript.PayToAddrScript(dAddr)
	if err != nil {
		return err
	}
	txOut := wire.NewTxOut(amountInSatoshi, pubScript)
	tx.AddTxOut(txOut)

	// create change output if any
	fee := int64(1000)
	changeAmount := totalAmount - amountInSatoshi - fee

	if changeAmount > 0 {
		changeScript, err := txscript.PayToAddrScript(myAddress)
		if err != nil {
			return nil
		}
		out := wire.NewTxOut(changeAmount, changeScript)
		tx.AddTxOut(out)
	}

	//create witness
	for i, v := range usedInputs {

		privKey := getPrivKey(uint32(sender.Index))

		// create witness for this input
		scriptPubKeyBytes, _ := hex.DecodeString(v.ScriptPubKey)

		fet := txscript.NewCannedPrevOutputFetcher(scriptPubKeyBytes, v.AmountInSatoshi())
		txSig := txscript.NewTxSigHashes(tx, fet)
		witness, err := txscript.WitnessSignature(tx, txSig, i, v.AmountInSatoshi(), scriptPubKeyBytes, txscript.SigHashAll, privKey, true)
		if err != nil {
			slog.Error("Error creating witness signature", "error", err)
			return err
		}

		tx.TxIn[i].Witness = witness
		engine, err := txscript.NewEngine(scriptPubKeyBytes, tx, i, txscript.StandardVerifyFlags, nil, nil, v.AmountInSatoshi(), fet)
		if err != nil {
			slog.Error("Error Creating engine,", "Error", err)
			return err
		}
		if err := engine.Execute(); err != nil {
			slog.Error("Error executing transaction script", "error", err)
			return err
		}
	}
	buf := new(bytes.Buffer)
	tx.Serialize(buf)
	trxHex := hex.EncodeToString(buf.Bytes())
	fmt.Println(trxHex)
	fmt.Println("Transaction is valid")

	// send the transaction
	fmt.Println("Sending transaction...")
	msg, err := blockchain.QueryFromBytes("sendrawtransaction", []byte(trxHex))
	if err != nil {
		slog.Error("Error sending raw transaction", "error", err)
		return err

	}
	fmt.Println("Transaction Sent!!!")
	json.NewEncoder(os.Stdout).Encode(msg)

	return nil
}

func getPrivKey(addressIndex uint32) *btcec.PrivateKey {
	// private key for account id 17
	keyStr := "tprv8fQcvSh37DJP7fxSKvJKyZHxCsVX5m9tGpcM21H3WuBYnQERJhU8bhPEzDtanzkaPA9han5cxMt6PXxbqkqRKUMvGnKceQYuFzfHru15667"
	extKey, _ := hdkeychain.NewKeyFromString(keyStr)
	changeKey, _ := extKey.Derive(0)

	childKey, _ := changeKey.Derive((addressIndex))
	privKey, _ := childKey.ECPrivKey()
	return privKey
}
