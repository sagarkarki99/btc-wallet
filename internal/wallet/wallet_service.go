package wallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/btcsuite/btcd/btcutil"
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
	addr, _ := ws.kc.GenerateAddress()
	payload, err := ws.getDescriptorPayload(addr)

	_, e := blockchain.QueryFromBytes("importdescriptors", payload)
	if e != nil {
		slog.Error("Error importing address", "error", e)
	}

	if err != nil {
		fmt.Println("Error getting wallet : ", err)
	}

	w := db.Wallet{
		Address: addr.String(),
		UserId:  userId,
	}
	ws.repo.Save(w)
	return w.Address
}

func (ws *WalletServiceImpl) getDescriptorPayload(addr *btcutil.AddressPubKey) ([]byte, error) {
	uncompressedPubKey := addr.PubKey().SerializeUncompressed()
	compressedPubKey := addr.PubKey().SerializeCompressed()
	slog.Info("Key information",
		"UncompressedPubKey", hex.EncodeToString(uncompressedPubKey),
		"UncompressedByte", len(uncompressedPubKey),
		"compressedByte", len(compressedPubKey),
		"CompressedPubKey", hex.EncodeToString(compressedPubKey),
		"Address", addr.AddressPubKeyHash().String(),
		"EncodedAddress", addr.EncodeAddress(),
	)
	data, err := blockchain.Query("getdescriptorinfo", []interface{}{"addr(" + addr.EncodeAddress() + ")"})
	if err != nil {
		return nil, err
	}
	var dataMap map[string]interface{}
	r, _ := data.MarshalJSON()

	json.Unmarshal(r, &dataMap)
	des := fmt.Sprintf("addr(%s)#%s", addr, dataMap["checksum"].(string))
	payload := map[string]interface{}{
		"desc":      des,
		"timestamp": "now",
		"watchonly": true,
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
	utxo, err := ws.getUTXOs(sender.Address)
	if err != nil {
		fmt.Println("Error getting UTXOs: ", err)
		return err
	}

	if utxo.TotalAmount < amount {
		return fmt.Errorf("insufficient funds")
	}

	// var tx wire.MsgTx

	json.NewEncoder(os.Stdout).Encode(utxo)

	// hash, err := chainhash.NewHashFromStr(utxo.Inputs[1].Txid)
	// if err != nil {
	// 	return fmt.Errorf("error creating hash: %v", err)
	// }
	// outputPoint := wire.NewOutPoint(hash, uint32(utxo.Inputs[1].Vout))
	// tx.AddTxIn(wire.NewTxIn(outputPoint, nil, nil))

	// if err != nil {
	// 	return fmt.Errorf("invalid address: %v", err)
	// }

	// address, err := btcutil.DecodeAddress(sender.Address, &chaincfg.RegressionNetParams)
	// if err != nil {
	// 	return fmt.Errorf("invalid address: %v", err)
	// }
	// pkScript, err := txscript.PayToAddrScript(address)
	// if err != nil {
	// 	return fmt.Errorf("error creating script: %v", err)
	// }
	// satoshis, err := btcutil.NewAmount(amount)
	// if err != nil {
	// 	return fmt.Errorf("error creating amount: %v", err)
	// }

	// txOut := wire.NewTxOut(int64(satoshis), pkScript)
	// tx.AddTxOut(txOut)

	// fmt.Printf("Transaction Details:\n")
	// fmt.Printf("  Version: %d\n", tx.Version)
	// fmt.Printf("  Inputs (%d):\n", len(tx.TxIn))
	// for i, in := range tx.TxIn {
	// 	fmt.Printf("    Input %d:\n", i)
	// 	fmt.Printf("      PrevTxHash: %s\n", in.PreviousOutPoint.Hash.String())
	// 	fmt.Printf("      PrevTxIndex: %d\n", in.PreviousOutPoint.Index)
	// 	fmt.Printf("      Sequence: %d\n", in.Sequence)
	// }
	// fmt.Printf("  Outputs (%d):\n", len(tx.TxOut))
	// for i, out := range tx.TxOut {
	// 	fmt.Printf("    Output %d:\n", i)
	// 	fmt.Printf("      Value: %d satoshis\n", out.Value)
	// 	fmt.Printf("      Script : %s\n", out.PkScript)
	// }
	// fmt.Printf("  LockTime: %d\n", tx.LockTime)
	// tx := wire.NewMsgTx(1)
	// tx.AddTxIn(&wire.NewTxIn())
	// Get UTXOs from the address. send this tx to keychain and keychain will add signature to this payload.

	// create transaction payload
	// Sign it using keychain
	// broad cast it to the network.
	return nil
}
