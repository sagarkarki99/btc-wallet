package main

import (
	"fmt"

	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/keychain"
	"github.com/sagarkarki99/internal/wallet"
)

func main() {
	db.Connect()
	RunApp()

}

func RunApp() {
	fmt.Println("Welcome to your bitcoin wallet., ")
	kc := keychain.NewKeychain()
	ws := wallet.NewWalletService(kc)
	addr := ws.GetDepositAddress("1")
	fmt.Println("Deposit Address: ", addr)
	fmt.Scanf("%s", &addr)
}
