package wallet

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/blockchain"
	"github.com/sagarkarki99/internal/keychain"
	repo "github.com/sagarkarki99/internal/repository"
)

type WalletService interface {
	GetDepositAddress(userId string) string
	GetBalance(userId string) float64
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
		Address: addr,
		UserId:  userId,
	}
	ws.repo.Save(w)
	return w.Address
}

func (ws *WalletServiceImpl) getDescriptorPayload(addr string) ([]byte, error) {
	data, err := blockchain.Query("getdescriptorinfo", []interface{}{"addr(" + addr + ")"})
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
	utxoMap, err := ws.getUTXOs(addr)
	if err != nil {
		fmt.Println("Error getting UTXOs: ", err)
		return 0
	}

	if total, ok := utxoMap["total_amount"].(float64); ok {
		return total
	}
	fmt.Println("Total Amount: ", utxoMap)
	return 0

}

func (*WalletServiceImpl) getUTXOs(addr string) (map[string]interface{}, error) {
	address := "addr(" + addr + ")"
	params := []interface{}{"start", []string{address}}
	res, _ := blockchain.Query("scantxoutset", params)
	r, _ := res.MarshalJSON()
	var utxoMap map[string]interface{}
	if err := json.Unmarshal(r, &utxoMap); err != nil {
		fmt.Println("Error unmarshaling UTXO response:", err)
		return nil, err
	}
	return utxoMap, nil
}

func (ws *WalletServiceImpl) SendToAddress(userId string, amount float64, destinationAddress string) error {
	sender, _ := ws.repo.Get(userId)
	utxos, err := ws.getUTXOs(sender.Address)
	if err != nil {
		fmt.Println("Error getting UTXOs: ", err)
		return err
	}

	if utxos["total_amount"].(float64) < amount {
		return fmt.Errorf("insufficient funds")
	}

	tx := btcutil.NewTx(wire.NewMsgTx(1))

	// Get UTXOs from the address. send this tx to keychain and keychain will add signature to this payload.
	tx.MsgTx().TxIn = []*wire.TxIn{}
	tx.MsgTx().TxOut = []*wire.TxOut{}
	// create transaction payload
	// Sign it using keychain
	// broad cast it to the network.
	return nil
}
