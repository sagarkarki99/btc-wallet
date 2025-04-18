package wallet

import (
	"encoding/json"
	"fmt"

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
	wallet, err := ws.repo.Get(userId)
	if wallet != nil {
		return wallet.Address
	}
	addr, _ := ws.kc.GenerateAddress()

	w := db.Wallet{
		Address: addr,
		UserId:  userId,
	}
	ws.repo.Save(w)
	if err != nil {
		fmt.Println("Error getting wallet : ", err)
	}
	return w.Address
}

func (ws *WalletServiceImpl) GetBalance(addr string) float64 {
	fmt.Println("Getting UTXOs for address:... ", addr)
	fmt.Println("-------------")
	address := "addr(" + addr + ")"
	params := []interface{}{"start", []string{address}}
	fmt.Println(params)
	res, _ := blockchain.Query("scantxoutset", params)
	r, _ := res.MarshalJSON()
	var utxoMap map[string]interface{}
	if err := json.Unmarshal(r, &utxoMap); err != nil {
		fmt.Println("Error unmarshaling UTXO response:", err)
		return 0
	}

	if total, ok := utxoMap["total_amount"].(float64); ok {
		return total
	}
	fmt.Println("Total Amount: ", utxoMap)
	return 0

}

func (ws *WalletServiceImpl) SendToAddress(userId string, amount float64, destinationAddress string) error {
	// create transaction payload
	// Sign it using keychain
	// broad cast it to the network.
	return nil
}
