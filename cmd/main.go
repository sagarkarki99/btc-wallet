package main

import (
	"fmt"

	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/blockchain"
	"github.com/sagarkarki99/internal/keychain"
	"github.com/sagarkarki99/internal/wallet"
)

func main() {
	db.Connect()
	go blockchain.Start()
	RunApp()

}

func RunApp() {
	fmt.Println("Welcome to your bitcoin wallet., ")
	kc := keychain.NewKeychain()
	ws := wallet.NewWalletService(kc)
	addr := ws.GetDepositAddress("11")
	fmt.Println("Deposit Address: ", addr)
	balance := ws.GetBalance(addr)
	fmt.Println("Deposit Address: ", addr)
	fmt.Println("Balance: ", balance)
	for {
	}
}
