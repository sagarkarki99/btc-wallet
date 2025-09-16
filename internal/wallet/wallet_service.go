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
	CreateWallet() string
	GetDepositAddress(userId int) string
	GetBalance(userId int) float64
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

func (ws *WalletServiceImpl) CreateWallet() string {
	ai, err := ws.kc.CreateAccount()
	if err != nil {
		return ""
	}

	payload, _ := ws.getDescriptorPayload(ai.Fingerprint, ai.Xpub)

	_, e := blockchain.Query("importdescriptors", []interface{}{payload})
	if e != nil {
		slog.Error("Error importing address", "error", e)
	}

	addr := generateAddress(ai.Xpub, ai.Id, 0)
	ws.repo.SaveWallet(ai.Xpub, ai.Id, ai.Fingerprint)
	ws.repo.SaveAddress(*addr)
	return addr.Hash
}

func (ws *WalletServiceImpl) GetDepositAddress(userId int) string {
	wallet, _ := ws.repo.GetWalletInfo(userId)
	if wallet == nil {
		return ""
	}

	address := generateAddress(wallet.XpubKey, wallet.AccountId, wallet.NextAddressIndex)
	ws.repo.SaveAddress(*address)

	return address.Hash
}

func generateAddress(xpub string, accountId, addressIndex int) *db.Address {
	extKey, _ := hdkeychain.NewKeyFromString(xpub)

	// We received m/84h/0h/0h from addressInfo.
	// Adding change type (0 for external, 1 for internal)
	//Path: m/84h/0h/0h/0
	changeKey, _ := extKey.Derive(0)

	// Adding index for address generation (0,1,2...)
	//Path: m/84h/0h/0h/0/0

	// we need to import m/84h/0h/0h/xpub/0/* to the node.
	childKey, err := changeKey.Derive(uint32(addressIndex))
	if err != nil {
		slog.Error("Error deriving child key", "error", err)
	}

	childPub, _ := childKey.ECPubKey()

	addr, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(childPub.SerializeCompressed()), blockchain.GetNetworkParams())
	address := &db.Address{
		Hash:      addr.EncodeAddress(),
		Index:     addressIndex,
		NextIndex: addressIndex + 1,
		AccountId: accountId,
	}
	fmt.Printf("Address ( %d): %s", address.Index, addr.EncodeAddress())
	fmt.Println("--------------------------------")
	return address
}

func (ws *WalletServiceImpl) getDescriptorPayload(fp string, xpub string) (interface{}, error) {

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
	return []interface{}{payload}, err
}

func (ws *WalletServiceImpl) GetBalance(userId int) float64 {
	addr, err := ws.repo.GetAddress(userId)
	if err != nil {
		return 0
	}
	utxo, err := ws.getUTXOs(addr.Hash)
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

		privKey := getPrivKey(uint32(sender.AccountId), uint32(sender.Index))

		// create witness for this input
		scriptPubKeyBytes, _ := hex.DecodeString(v.ScriptPubKey)

		// =================================================================
		// START: DIAGNOSTIC BLOCK
		// =================================================================
		fmt.Printf("\n--- [DEBUGGING INPUT %d] ---\n", i)
		fmt.Printf("UTXO Txid: %s:%d\n", v.Txid, v.Vout)

		// Step 1: Derive the Public Key Hash from YOUR private key.
		// We assume compressed, as this is the standard for SegWit.
		myPublicKey := privKey.PubKey().SerializeCompressed()
		myPubKeyHash := btcutil.Hash160(myPublicKey)
		fmt.Printf("Hash derived from YOUR private key: %x\n", myPubKeyHash)

		// Step 2: Extract the required Public Key Hash from the UTXO's script.
		// A P2WPKH scriptPubKey is 0x0014{20-byte-hash}. The first two bytes are the witness version and push opcode.
		if len(scriptPubKeyBytes) != 22 || scriptPubKeyBytes[0] != 0x00 || scriptPubKeyBytes[1] != 0x14 {
			slog.Error("This does not appear to be a standard P2WPKH scriptPubKey!", "script", v.ScriptPubKey)
			return errors.New("invalid scriptpubkey format for P2WPKH")
		}
		requiredPubKeyHash := scriptPubKeyBytes[2:]
		fmt.Printf("Hash required by the UTXO script:   %x\n", requiredPubKeyHash)

		// Step 3: Compare them. This is the check that OP_EQUALVERIFY performs.
		if !bytes.Equal(myPubKeyHash, requiredPubKeyHash) {
			slog.Error("!!! KEY MISMATCH !!! The private key you are using does not correspond to the UTXO being spent.")
			fmt.Println("This is why OP_EQUALVERIFY is failing.")
			return errors.New("key mismatch: provided private key cannot spend this UTXO")
		} else {
			fmt.Println("✅ Key/Hash Match Confirmed. The private key is correct for this UTXO.")
		}
		// =================================================================
		// END: DIAGNOSTIC BLOCK
		// =================================================================

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
	msg, err := blockchain.Query("sendrawtransaction", []interface{}{trxHex})
	if err != nil {
		slog.Error("Error sending raw transaction", "error", err)
		return err

	}
	fmt.Println("Transaction Sent!!!")
	json.NewEncoder(os.Stdout).Encode(msg)

	return nil
}

func getPrivKey(accountId, addressIndex uint32) *btcec.PrivateKey {
	keyStr := "tprv8fLoca4ervKAYJ9f5b2jGTyZa2Sqj81VBjJXBx9XSQKUkThCYdcRGwqTNAt5eQeADPmPQwvB2K3CJeCutWaKKJFCtQUzQL6y5fVf2TTUdem"
	extKey, _ := hdkeychain.NewKeyFromString(keyStr)
	changeKey, _ := extKey.Derive(0)

	childKey, _ := changeKey.Derive((addressIndex))
	privKey, _ := childKey.ECPrivKey()
	childPub, _ := childKey.ECPubKey()

	addr, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(childPub.SerializeCompressed()), blockchain.GetNetworkParams())
	fmt.Printf("Address ( m/84h/1h/%dh/0/%d): %s", accountId, addressIndex, addr.EncodeAddress())
	fmt.Println("--------------------------------")

	return privKey
}
