package main

import (
	"fmt"

	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/wallet"
)

func main() {
	db.Connect()
	fmt.Println("Welcome to bitcoin wallet server")
	ws := wallet.NewWalletService()
	addr := ws.GetDepositAddress("1")
	fmt.Println("Deposit Address: ", addr)
}
