package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
)

type WalletService interface {
	Create() string
}

func NewWalletService() WalletService {
	return &WalletServiceImpl{}
}

type WalletServiceImpl struct{}

func (ws *WalletServiceImpl) Create() string {

	pk, err := ws.generatePrivateKey()
	if err != nil {
		fmt.Printf("error generating private key: %v", err)
		return ""
	}

	pubKey, err := ws.generatePublicKey(pk)
	if err != nil {
		fmt.Printf("error generating public key: %v", err)
		return ""
	}
	fmt.Println(pubKey)

	// generate public key from
	// encrypt it afterwards to save to db

	return ""
}

func (ws *WalletServiceImpl) generatePublicKey(pk []byte) (string, error) {
	_, pubKey := btcec.PrivKeyFromBytes(pk)

	return hex.EncodeToString(pubKey.SerializeCompressed()), nil
}

func (ws *WalletServiceImpl) generatePrivateKey() ([]byte, error) {
	// generate random private key
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("error generating random bytes: %v", err)
	}

	fmt.Println("Raw Private Key : ", hex.EncodeToString(privateKey))

	return privateKey, nil
}
