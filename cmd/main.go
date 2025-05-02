package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/wallet/deposit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			UserId string `json:"userId"`
		}

		json.NewDecoder(r.Body).Decode(&req)
		if req.UserId == "" {
			http.Error(w, "UserId is required", http.StatusBadRequest)
			return
		}

		addr := ws.GetDepositAddress(req.UserId)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"address": addr,
		}
		responseBytes, _ := json.Marshal(response)
		w.Write(responseBytes)
		defer r.Body.Close()

	})

	r.HandleFunc("/wallet/GetBalance", func(w http.ResponseWriter, r *http.Request) {
		addr := ws.GetDepositAddress("11")
		ws.GetBalance(addr)
		res := map[string]float64{
			"balance": balance,
		}
		resBytes, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resBytes)
	})

	r.HandleFunc("/api/v1/wallet/send", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var requestBody struct {
			UserId        string  `json:"userId"`
			Amount        float64 `json:"amount"`
			SenderAddress string  `json:"senderAddress"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if requestBody.UserId == "" || requestBody.Amount <= 0 || requestBody.SenderAddress == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
		// w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(http.StatusOK)
		// by, _ := json.Marshal(requestBody)
		// w.Write(by)

		if err := ws.SendToAddress(requestBody.UserId, requestBody.Amount, requestBody.SenderAddress); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Failed to send transaction",error: "` + err.Error() + `"}`))
		}
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
