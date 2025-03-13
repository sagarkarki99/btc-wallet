package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
)

type TransactionService interface {
	Create()
	GetTransaction()
}

func NewTransactionService(ws WalletService) TransactionService {
	connConfig := rpcclient.ConnConfig{
		Host:         "host.docker.internal:18332",
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rpcClient, err := rpcclient.New(&connConfig, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &transactionService{
		ws:        ws,
		rpcClient: rpcClient,
	}
}

type transactionService struct {
	ws        WalletService
	rpcClient *rpcclient.Client
}

func (s *transactionService) Create() {
	// create a transaction from here.
	// tx := btcutil.NewTx(wire.NewMsgTx(wire.TxVersion))
}

func (s *transactionService) GetTransaction() {
	// Example: Get blockchain info
	// blockchainInfo, err := s.rpcClient.GetBlockChainInfo()
	// log.Println("blockchainInfo", blockchainInfo)

	// if err != nil {
	// 	log.Printf("Error getting blockchain info: %v", err)
	// 	return
	// }

	// fmt.Println(blockchainInfo)
	// // Example: Scan for specific address
	// cmd := "getblockchaininfo"
	// params := []interface{}{
	// 	"start",
	// 	[]string{"addr(tb1qd5qt4e7dwtjn8s8smrtgyxtkazpcj5get02jyr)"},
	// }
	res := make(map[string]interface{})
	response, err := s.rpcClient.RawRequest("getblockchaininfo", nil)

	if err != nil {
		log.Printf("Error getting blockchain info: %v", err)
		return
	}
	json.Unmarshal(response, &res)
	fmt.Println("BlockchainInfo:", res)
	// when running bitcoind -daemon, it will run in 120.0.0.1:28842
	// Connect to 120.0.0.1:28842 that is running in this machine (not in container)
	// make a rpc call to scantxoutset start '["addr(tb1qd5qt4e7dwtjn8s8smrtgyxtkazpcj5get02jyr)"]'
}
