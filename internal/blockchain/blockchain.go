package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
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
			slog.Error("Failed to connect to Bitcoin node: %v", "Error", err.Error())
			slog.Error("Please check if your Bitcoin node is running and accessible from Docker")
			slog.Error("Ensure 'rpcallowip' in bitcoin.conf allows connections from Docker networks")
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
		Host:         fmt.Sprintf("%s:%s", host, port),
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rpcManager = &RPCClientManager{
		config: &connConfig,
		endpoints: map[string]string{
			"mywallets":         "mywallets",
			"descriptorwallets": "descriptorwallets",
		},
	}

	response, err := rpcManager.GetClient("").RawRequest("getblockchaininfo", nil)
	if err != nil {
		log.Printf("Error getting blockchain info: %v", err)
		return
	}
	jsonbyte, _ := response.MarshalJSON()
	jsonData := string(jsonbyte)
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(jsonData)
	go listenToNode()
}

func QueryFromBytes(rpcMethod string, data []byte) (*json.RawMessage, error) {

	res, err := rpcManager.GetClient("descriptorwallet").RawRequest(rpcMethod, []json.RawMessage{data})
	return &res, err
}

func Query(rpcMethod string, params []interface{}) (*json.RawMessage, error) {
	jsonParams := make([]json.RawMessage, len(params))
	for i, param := range params {
		jsonBytes, err := json.Marshal(param)
		if err != nil {
			slog.Error("Error marshaling parameter", "parameter", param, "error", err.Error())
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
	godotenv.Load(".env")
	host := os.Getenv("BTC_RPC_NODE")
	err = sub.Connect(fmt.Sprintf("tcp://%s", host))
	fmt.Println("Connected to socket: ", sub)
	if err != nil {
		slog.Error("Could not connect to socket: ", "Error", err.Error())
		// can add a retry logic here
		return
	}

	if err := sub.SetSubscribe("rawtx"); err != nil {
		fmt.Println("Could not set subscribe: ", err)
	}

	for {
		fmt.Println("Waiting for message...")
		msg, _ := sub.RecvMessageBytes(0)
		var trx wire.MsgTx

		if err := trx.Deserialize(bytes.NewReader(msg[1])); err != nil {
			slog.Error("Error deserializing transaction", "error", err.Error())
			return
		}

		jsonBytes, err := json.MarshalIndent(trx, "", "  ")
		if err != nil {
			slog.Error("Error marshaling transaction", "error", err.Error())
			return
		}
		fmt.Printf("Parsed Transaction:\n%s\n", string(jsonBytes))
	}
}
