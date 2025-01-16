package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/bech32"
	repo "github.com/sagarkarki99/internal/repository"
	"golang.org/x/crypto/ripemd160"
)

type WalletService interface {
	Create() string
}

func NewWalletService() WalletService {
	return &WalletServiceImpl{
		repo: repo.New(),
	}
}

type WalletServiceImpl struct {
	repo repo.WalletRepository
}

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

	// hash it with sha256
	hashed256 := sha256.Sum256(pubKey)

	// hash it with ripemd160
	ripeHasher := ripemd160.New()
	ripeHasher.Write(hashed256[:])
	hashedRIPEMD160 := ripeHasher.Sum(nil)

	modernAddress := segWitAddress(hashedRIPEMD160)
	fmt.Println("Modern Address (SegWit) : ", modernAddress)
	wallet := repo.Wallet{
		PrivateKey: hex.EncodeToString(pk),
		PublicKey:  hex.EncodeToString(pubKey),
		UserId:     "1",
	}
	id, err := ws.repo.Save(wallet)
	if err != nil {
		fmt.Println("Error saving wallet : ", err)
	}
	fmt.Println("Wallet ID: ", id)
	w, err := ws.repo.Get("1")
	if err != nil {
		fmt.Println("Error getting wallet : ", err)

	}
	fmt.Println("Wallet : ", w)
	return modernAddress
}

func segWitAddress(hashedRIPEMD160 []byte) string {

	bec32bytes, err := bech32.ConvertBits(hashedRIPEMD160, 8, 5, true)
	if err != nil {
		fmt.Println("Error converting bits : ", err)
	}
	bytesWithVersion := append([]byte{0}, bec32bytes...)
	address, _ := bech32.Encode("bc", bytesWithVersion)
	return address

}

func p2pkhAddress(hashedRIPEMD160 []byte) string {
	//versioning the hash
	versionedhash := append([]byte{0x00}, hashedRIPEMD160...)
	singleHashed := sha256.Sum256(versionedhash)
	doubleHashed := sha256.Sum256(singleHashed[:])

	// adding checksum
	firstFourBytes := doubleHashed[:4]
	finalHash := append(versionedhash, firstFourBytes...)
	fmt.Println("Final hash length : ", len(finalHash))

	// Encode with base58
	finalAddress := base58.Encode(finalHash)
	fmt.Println("Final Address : ", finalAddress)
	return finalAddress
}

func (ws *WalletServiceImpl) generatePublicKey(pk []byte) ([]byte, error) {
	_, pubKey := btcec.PrivKeyFromBytes(pk)

	return pubKey.SerializeCompressed(), nil
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
