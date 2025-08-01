package blockchain

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/joho/godotenv"
	"github.com/pebbe/zmq4"
)

type RPCClientManager struct {
	config *rpcclient.ConnConfig
	Host   string
	Port   string
}

var activeWallet = "external_wallet"

func (cm *RPCClientManager) GetClient(walletname string) *rpcclient.Client {
	if activeWallet == "" {
		conn, err := rpcclient.New(cm.config, nil)
		if err != nil {
			slog.Error("Failed to connect to Bitcoin node: %v", "Error", err.Error())
			slog.Error("Please check if your Bitcoin node is running and accessible from Docker")
			slog.Error("Ensure 'rpcallowip' in bitcoin.conf allows connections from Docker networks")
		}
		return conn
	}
	config := cm.config
	config.Host = fmt.Sprintf("%s:%s/wallet/%s", cm.Host, cm.Port, activeWallet)
	slog.Info("Config", "Host", config.Host)
	conn, _ := rpcclient.New(config, nil)
	return conn

}

var rpcManager *RPCClientManager

func Start(ctx context.Context) {
	// Connect to the node from here and start..
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("BTC_RPC_NODE")
	port := os.Getenv("BTC_RPC_PORT")
	user := os.Getenv("BTC_RPC_USER")
	pass := os.Getenv("BTC_RPC_PASS")

	connConfig := rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s:%s", host, port),
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rpcManager = &RPCClientManager{
		config: &connConfig,
		Host:   host,
		Port:   port,
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
	go listenToNode(ctx)
}

func QueryFromBytes(rpcMethod string, data []byte) (*json.RawMessage, error) {

	res, err := rpcManager.GetClient(activeWallet).RawRequest(rpcMethod, []json.RawMessage{data})
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
	res, err := rpcManager.GetClient(activeWallet).RawRequest(rpcMethod, jsonParams)
	return &res, err
}

func listenToNode(ctx context.Context) {
	sub, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		fmt.Println("Could not create socket: ", err)
	}
	defer sub.Close()
	godotenv.Load(".env")
	host := os.Getenv("BTC_RPC_NODE")
	err = sub.Connect(fmt.Sprintf("tcp://%s:28332", host))
	if err != nil {
		slog.Error("Could not connect to socket: ", "Error", err.Error())
		return
	}
	fmt.Println("Connected to socket: ", sub)

	if err := sub.SetSubscribe("hashblock"); err != nil {
		fmt.Println("Could not set subscribe: ", err)
	}
	if err := sub.SetRcvtimeo(1 * time.Second); err != nil {
		slog.Error("Could not set receive timeout: ", "Error", err.Error())
		return
	}

	fmt.Println("Waiting for message...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done, exiting listener")
			return
		default:
			msg, err := sub.RecvMessageBytes(0)
			if err != nil {
				continue
			}
			var block wire.MsgBlock

			if err := block.Deserialize(bytes.NewReader(msg[1])); err != nil {
				slog.Error("Error deserializing transaction", "error", err.Error())

			}
			_, err = os.Stat("transaction.json")
			if err != nil {
				if os.IsNotExist(err) {
					os.Create("transaction.json")
				}
			}
			f, _ := os.OpenFile("transaction.json", os.O_RDWR|os.O_APPEND, 0644)
			defer f.Close()
			var d []byte
			f.Read(d)

			io := bufio.NewWriter(f)
			json.NewEncoder(io).Encode(&block)

			io.WriteString(",")
			io.Flush()

			jsonBytes, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				slog.Error("Error marshaling transaction", "error", err.Error())
				return
			}
			fmt.Printf("Parsed Transaction:\n%s\n", string(jsonBytes))

		}
	}
}
