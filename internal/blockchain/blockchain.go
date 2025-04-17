package blockchain

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
)

var rpcClient *rpcclient.Client

func Start() {
	// Connect to the node from here and start..
	connConfig := rpcclient.ConnConfig{
		Host:         "host.docker.internal:18443",
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rc, err := rpcclient.New(&connConfig, nil)

	if err != nil {
		log.Printf("Failed to connect to Bitcoin node: %v", err)
		log.Printf("Please check if your Bitcoin node is running and accessible from Docker")
		log.Printf("Ensure 'rpcallowip' in bitcoin.conf allows connections from Docker networks")
		log.Fatal(err)
	}
	rpcClient = rc

	response, err := rpcClient.RawRequest("getblockchaininfo", nil)
	if err != nil {
		log.Printf("Error getting blockchain info: %v", err)
		return
	}
	jsonbyte, _ := response.MarshalJSON()
	jsonData := string(jsonbyte)
	fmt.Println("BlockchainInfo:", jsonData)

	// when running bitcoind -daemon, it will run in 120.0.0.1:28842
	// Connect to 120.0.0.1:28842 that is running in this machine (not in container)
	// make a rpc call to scantxoutset start '["addr(tb1qd5qt4e7dwtjn8s8smrtgyxtkazpcj5get02jyr)"]'
}

func Query(rpcMethod string, params []interface{}) (*json.RawMessage, error) {
	jsonParams := make([]json.RawMessage, len(params))
	for i, param := range params {
		jsonBytes, err := json.Marshal(param)
		if err != nil {
			log.Printf("Error marshaling parameter %d: %v", i, err)
			return nil, err
		}
		jsonParams[i] = jsonBytes
	}
	res, err := rpcClient.RawRequest(rpcMethod, jsonParams)
	return &res, err
}
