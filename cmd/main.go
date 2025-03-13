package main

import (
	"fmt"

	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/services"
)

func main() {
	db.Connect()
	fmt.Println("Welcome to bitcoin wallet server")
	ws := services.NewWalletService()
	ts := services.NewTransactionService(ws)

	// ws.Create()
	// fmt.Println(ws.Get("3"))
	ts.GetTransaction()
}
