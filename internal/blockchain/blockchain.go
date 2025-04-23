package blockchain

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/pebbe/zmq4"
)

type RPCClientManager struct {
	config    *rpcclient.ConnConfig
	endpoints map[string]string
}

func (cm *RPCClientManager) GetClient(walletname string) *rpcclient.Client {
	if cm.endpoints[walletname] == "" {
		conn, err := rpcclient.New(cm.config, nil)
		if err != nil {
			log.Printf("Failed to connect to Bitcoin node: %v", err)
			log.Printf("Please check if your Bitcoin node is running and accessible from Docker")
			log.Printf("Ensure 'rpcallowip' in bitcoin.conf allows connections from Docker networks")
		}
		return conn
	}
	config := cm.config
	config.Host = fmt.Sprintf("%s/wallet/%s", cm.config.Host, cm.endpoints[walletname])
	conn, _ := rpcclient.New(config, nil)
	return conn

}

var rpcManager *RPCClientManager

func Start() {
	// Connect to the node from here and start..
	connConfig := rpcclient.ConnConfig{
		Host:         "host.docker.internal:18443",
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rpcManager = &RPCClientManager{
		config: &connConfig,
	}

	response, err := rpcManager.GetClient("").RawRequest("getblockchaininfo", nil)
	if err != nil {
		log.Printf("Error getting blockchain info: %v", err)
		return
	}
	jsonbyte, _ := response.MarshalJSON()
	jsonData := string(jsonbyte)
	fmt.Println("BlockchainInfo:", jsonData)
	go listenToNode()
	// when running bitcoind -daemon, it will run in 120.0.0.1:28842
	// Connect to 120.0.0.1:28842 that is running in this machine (not in container)
	// make a rpc call to scantxoutset start '["addr(tb1qd5qt4e7dwtjn8s8smrtgyxtkazpcj5get02jyr)"]'
}

func QueryFromBytes(rpcMethod string, data []byte) (*json.RawMessage, error) {

	res, err := rpcManager.GetClient("nokeyswallet").RawRequest(rpcMethod, []json.RawMessage{data})
	return &res, err
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
	res, err := rpcManager.GetClient("").RawRequest(rpcMethod, jsonParams)
	return &res, err
}

func listenToNode() {
	sub, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		fmt.Println("Could not create socket: ", err)
	}
	defer sub.Close()

	err = sub.Connect("tcp://host.docker.internal:28332")
	fmt.Println("Connected to socket: ", sub)
	if err != nil {
		fmt.Println("Could not connect to socket: ", err)
	}

	if err := sub.SetSubscribe("hashblock"); err != nil {
		fmt.Println("Could not set subscribe: ", err)
	}

	for {
		fmt.Println("Waiting for message...")
		msg, _ := sub.RecvMessageBytes(0)

		fmt.Println("Parsed Message with Sprintf: ", fmt.Sprintf("%x", msg[1]))
	}
}
