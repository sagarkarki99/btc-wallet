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

type Utxo struct {
	Inputs      []Input `json:"unspents"`
	TotalAmount float64 `json:"total_amount"`
}

type WalletService interface {
	GetDepositAddress(userId string) string
	GetBalance(addr string) float64
	SendToAddress(userId string, amount float64, destinationAddress string) error
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

func (ws *WalletServiceImpl) GetDepositAddress(userId string) string {

	wallet, _ := ws.repo.Get(userId)
	if wallet != nil {
		return wallet.Address
	}

	// This will be an accountId (can be based on userId )
	addrInfo, _ := ws.kc.GenerateAddress(0)

	payload, _ := ws.getDescriptorPayload(addrInfo)

	_, e := blockchain.QueryFromBytes("importdescriptors", payload)
	if e != nil {
		slog.Error("Error importing address", "error", e)
	}

	extKey, _ := hdkeychain.NewKeyFromString(addrInfo.Xpub)

	// We received m/84h/0h/0h from addressInfo.
	// Adding change type (0 for external, 1 for internal)
	//Path: m/84h/0h/0h/0
	changeKey, _ := extKey.Derive(0)

	// Adding index for address generation (0,1,2...)
	//Path: m/84h/0h/0h/0/0

	// we need to import m/84h/0h/0h/xpub/0/* to the node.
	childKey, err := changeKey.Derive(0)
	if err != nil {
		slog.Error("Error deriving child key", "error", err)
	}

	pKey, _ := childKey.ECPubKey()
	address, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pKey.SerializeCompressed()), blockchain.GetNetworkParams())

	childPub, _ := childKey.ECPubKey()

	// Address at 0 index
	addr, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(childPub.SerializeCompressed()), blockchain.GetNetworkParams())
	fmt.Println("Address ( m/84h/1h/0h/0/0): ", addr.EncodeAddress())
	fmt.Println("--------------------------------")
	childKey1, _ := changeKey.Derive(1)

	childPub1, _ := childKey1.ECPubKey()
	fmt.Println("Child Key Public Key:", hex.EncodeToString(childPub1.SerializeCompressed()))
	fmt.Println("Child Key Public Key EXTENDED:", childKey1.String())
	addr2, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(childPub1.SerializeCompressed()), blockchain.GetNetworkParams())
	fmt.Println("Address ( m/84h/1h/0h/0/1): ", addr2.EncodeAddress())

	w := db.Wallet{
		Address: address.EncodeAddress(),
		UserId:  userId,
	}
	ws.repo.Save(w)
	return w.Address
}

func (ws *WalletServiceImpl) getDescriptorPayload(addr *keychain.AddressInfo) ([]byte, error) {

	p := fmt.Sprintf("wpkh([%s/84h/1h/0h]%s/0/*)", addr.Fingerprint, addr.Xpub)
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

func (ws *WalletServiceImpl) SendToAddress(userId string, amount float64, destinationAddress string) error {
	sender, err := ws.repo.Get(userId)
	if err != nil {
		return errors.New("user do not have any wallet")
	}
	fmt.Println("Sender Address: ", sender.Address)
	u, err := ws.getUTXOs(sender.Address)
	if err != nil {
		fmt.Println("Error getting UTXOs: ", err)
		return err
	}

	if u.TotalAmount < amount {
		return fmt.Errorf("insufficient funds")
	}

	utxo := u.Inputs[0]
	utxoAmountInSatoshi := int64(utxo.Amount * 1e8)
	amountInSatoshi := int64(amount * 1e8)

	tx := wire.NewMsgTx(wire.TxVersion)

	// create Input
	h, _ := chainhash.NewHashFromStr(utxo.Txid)
	op := wire.NewOutPoint(h, uint32(utxo.Vout))
	inp := wire.NewTxIn(op, []byte{}, nil)
	tx.AddTxIn(inp)

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
	changeAmount := utxoAmountInSatoshi - amountInSatoshi - fee
	changeAddr, _ := btcutil.DecodeAddress(sender.Address, blockchain.GetNetworkParams())
	if changeAmount > 0 {
		changeScript, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			return nil
		}
		out := wire.NewTxOut(changeAmount, changeScript)
		tx.AddTxOut(out)
	}

	//create witness
	myScriptPubKey, _ := txscript.PayToAddrScript(changeAddr)
	fetcher := newPrevOutputFetcher(utxo.Txid, uint32(utxo.Vout), utxoAmountInSatoshi, myScriptPubKey)
	txSig := txscript.NewTxSigHashes(tx, fetcher)
	witness, err := txscript.WitnessSignature(tx, txSig, 0, utxoAmountInSatoshi, myScriptPubKey, txscript.SigHashAll, getPrivKey(), true)
	if err != nil {
		slog.Error("Error creating witness signature", "error", err)
		return err
	}

	tx.TxIn[0].Witness = witness

	// validate the trx
	engine, _ := txscript.NewEngine(myScriptPubKey, tx, 0, txscript.StandardVerifyFlags, nil, nil, utxoAmountInSatoshi, fetcher)
	if err := engine.Execute(); err != nil {
		slog.Error("Error executing transaction script", "error", err)
		return err
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

func getPrivKey() *btcec.PrivateKey {
	keyStr := "tprv8fQcvSh37DJNwDVDSFomfqe1aTJB7gtJuDd87B6xEjJmAG5GsVcxRVJevfK59crfZh8WyRNq87okPhMZQgEvgQtrjtvXAyS7t8FDqHCv1ZY"
	extKey, _ := hdkeychain.NewKeyFromString(keyStr)
	changeKey, _ := extKey.Derive(0)

	childKey, _ := changeKey.Derive(0)
	privKey, _ := childKey.ECPrivKey()
	return privKey

}

type prevOutputFetcher struct {
	outputs map[wire.OutPoint]*wire.TxOut
}

func (f *prevOutputFetcher) FetchPrevOutput(op wire.OutPoint) *wire.TxOut {
	return f.outputs[op]
}

func newPrevOutputFetcher(txid string, vout uint32, amount int64, scriptPubKey []byte) *prevOutputFetcher {
	hash, _ := chainhash.NewHashFromStr(txid)
	op := wire.OutPoint{Hash: *hash, Index: vout}
	return &prevOutputFetcher{
		outputs: map[wire.OutPoint]*wire.TxOut{
			op: wire.NewTxOut(amount, scriptPubKey),
		},
	}
}
