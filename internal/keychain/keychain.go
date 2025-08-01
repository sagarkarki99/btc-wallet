package keychain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/sagarkarki99/db"
	repo "github.com/sagarkarki99/internal/repository"
)

type Keychain interface {
	GenerateAddress() (*btcutil.AddressPubKey, error)
	SignTransaction() string
}

type KeychainImpl struct {
	kr repo.KeychainRepository
}

func NewKeychain() Keychain {
	return &KeychainImpl{
		kr: repo.NewKeychainRepo(),
	}
}

func (kc *KeychainImpl) GenerateAddress() (*btcutil.AddressPubKey, error) {
	pk, _ := kc.generatePrivateKey()
	pubKey, _ := kc.generatePublicKey(pk)
	addr, err := btcutil.NewAddressPubKey(pubKey, &chaincfg.RegressionNetParams)
	if err != nil {
		fmt.Println("Error creating address : ", err)
		return nil, errors.New("error generating address")
	}

	kc.kr.Save(&db.KeyAddress{
		PrivateKey: hex.EncodeToString(pk),
		PublicKey:  hex.EncodeToString(pubKey),
	})

	return addr, nil
	// if err != nil {
	// 	fmt.Printf("error generating private key: %v", err)
	// 	return ""
	// }

	// pubKey, err := kc.generatePublicKey(pk)
	// if err != nil {
	// 	fmt.Printf("error generating public key: %v", err)
	// 	return ""
	// }
	// fmt.Println(pubKey)

	// // hash it with sha256
	// hashed256 := sha256.Sum256(pubKey)

	// // hash it with ripemd160
	// ripeHasher := ripemd160.New()
	// ripeHasher.Write(hashed256[:])
	// hashedRIPEMD160 := ripeHasher.Sum(nil)

	// modernAddress := segWitAddress(hashedRIPEMD160)

	// keys := &db.KeyAddress{
	// 	PrivateKey: hex.EncodeToString(pk),
	// 	PublicKey:  hex.EncodeToString(pubKey),
	// }
	// _, err = kc.kr.Save(keys)
	// if err != nil {
	// 	fmt.Println("Error saving wallet : ", err)
	// }
	// return modernAddress
}

func (kc *KeychainImpl) SignTransaction() string {
	// sign the transcation from here. Receive a transaction payload to this.
	return "signed transaction"
}

func segWitAddress(hashedRIPEMD160 []byte) string {

	bec32bytes, err := bech32.ConvertBits(hashedRIPEMD160, 8, 5, true)
	if err != nil {
		fmt.Println("Error converting bits : ", err)
	}
	bytesWithVersion := append([]byte{0}, bec32bytes...)
	address, _ := bech32.Encode("tb", bytesWithVersion)
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

func (kc *KeychainImpl) generatePublicKey(pk []byte) ([]byte, error) {
	_, pubKey := btcec.PrivKeyFromBytes(pk)

	return pubKey.SerializeCompressed(), nil
}

func (kc *KeychainImpl) generatePrivateKey() ([]byte, error) {
	// generate random private key
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("error generating random bytes: %v", err)
	}
	return privateKey, nil
}
