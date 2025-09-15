package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/blockchain"
	"github.com/sagarkarki99/internal/keychain"
	"github.com/sagarkarki99/internal/wallet"
)

var globalContext context.Context
var cancel context.CancelFunc

func main() {
	globalContext, cancel = context.WithCancel(context.Background())
	defer cancel()
	db.Connect()
	go blockchain.Start(globalContext)
	RunApp()

}

func RunApp() {
	fmt.Println("Welcome to your bitcoin wallet., ")
	kc := keychain.NewKeychain()
	ws := wallet.NewWalletService(kc)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/wallet/deposit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			UserId int `json:"userId"`
		}
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "UserId is required", http.StatusBadRequest)
			return
		}

		addr := ws.GetDepositAddress(req.UserId)
		response := map[string]string{
			"address": addr,
		}
		responseBytes, _ := json.Marshal(response)
		writeResponse(w, responseBytes, http.StatusOK)

	})
	r.HandleFunc("/api/v1/wallet/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		addr := ws.CreateWallet()
		response := map[string]string{
			"address": addr,
		}
		responseBytes, _ := json.Marshal(response)
		writeResponse(w, responseBytes, http.StatusOK)

	})

	r.HandleFunc("/api/v1/wallet/balance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
			return
		}

		params := r.URL.Query()
		userId := params.Get("userId")
		uId, _ := strconv.ParseInt(userId, 10, 32)
		balance := ws.GetBalance(int(uId))
		res := map[string]float64{
			"balance": balance,
		}
		resBytes, _ := json.Marshal(res)
		writeResponse(w, resBytes, http.StatusOK)
	})

	r.HandleFunc("/api/v1/wallet/send", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var requestBody struct {
			UserId        int     `json:"userId"`
			Amount        float64 `json:"amount"`
			SenderAddress string  `json:"senderAddress"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if requestBody.UserId == 0 || requestBody.Amount <= 0 || requestBody.SenderAddress == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		if err := ws.SendToAddress(requestBody.UserId, requestBody.Amount, requestBody.SenderAddress); err != nil {
			res := `{"message": "Failed to send transaction",error: "` + err.Error() + `"}`
			writeResponse(w, []byte(res), http.StatusBadRequest)
			return
		}

		writeResponse(w, []byte(`{"message":"Your transaction is sent to blockchain"}`), http.StatusOK)
	})

	slog.Info("Listening on port 8500")
	if err := http.ListenAndServe(":8500", r); err != nil {
		log.Fatal(err)
	}
}

func writeResponse(w http.ResponseWriter, data []byte, statusCode int) {
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}
