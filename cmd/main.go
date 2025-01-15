package main

import (
	"fmt"

	"github.com/sagarkarki99/internal/services"
)

func main() {
	fmt.Println("Welcome to bitcoin wallet server")
	ws := services.NewWalletService()
	ws.Create()
}
