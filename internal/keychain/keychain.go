package keychain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/blockchain"
	repo "github.com/sagarkarki99/internal/repository"
	"golang.org/x/crypto/ripemd160"
)

var (
	ErrGeneratingKey   = errors.New("error generating master key")
	ErrCreatingAccount = errors.New("error creating account key")
)

type Keychain interface {
	GenerateAddress(accountId uint32) (*AddressInfo, error)
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

func (kc *KeychainImpl) GenerateAddress(accountId uint32) (*AddressInfo, error) {
	masterKey, _ := kc.getMasterKey()
	mPubKey, _ := masterKey.Neuter() // Get the extended public key
	fmt.Println("Master Key Public Key:", mPubKey.String())

	fmt.Println("--------------------------")

	purposeKey, _ := masterKey.Derive(hdkeychain.HardenedKeyStart + 84)
	// Derive the coin type key for Bitcoin (0 for mainnet, 1 for testnet)
	// TODO: Refactor the hard coded coin type to be dynamic
	coinTypeKey, _ := purposeKey.Derive(hdkeychain.HardenedKeyStart + 0)

	// Path: m/84'/1'/0'
	// This is the account level. You would use a new index for each account (0, 1, 2...).
	// This is the extended private key you would store for the user.
	// Derive the account key (0 for the first account)
	// This should be incrmental.
	//TODO: Refactor the hard coded account index to be dynamic
	accountKey, err := coinTypeKey.Derive(hdkeychain.HardenedKeyStart + uint32(0))
	if err != nil {
		return nil, ErrGeneratingKey
	}

	// save xpriv key of accountKey
	// this is the account level keys: (private and public)
	// To sign the transactions for this account, extended private key is required.

	pvk, _ := accountKey.ECPrivKey()
	xpub, _ := accountKey.Neuter()
	fmt.Println("Is Private: ", accountKey.IsPrivate())
	fmt.Println("Account Key Private Key:", pvk.Key.String())
	fmt.Println("Account Key Public Key:", xpub.String())

	fp := kc.getMasterkeyFingerprint(masterKey)

	kc.kr.Save(&db.KeyAddress{
		PrivateKey: accountKey.String(),
		PublicKey:  xpub.String(),
	})

	return &AddressInfo{
		Fingerprint: fp,
		Xpub:        xpub.String(),
	}, nil

}

func (*KeychainImpl) getMasterkeyFingerprint(masterKey *hdkeychain.ExtendedKey) string {
	pub, _ := masterKey.ECPubKey()

	pubKeyBytes := pub.SerializeCompressed()
	shahash := sha256.Sum256(pubKeyBytes)
	ripeHasher := ripemd160.New()
	ripeHasher.Write(shahash[:])
	hashedRIPEMD160 := ripeHasher.Sum(nil)
	fingerprintBytes := hashedRIPEMD160[:4]
	fingerPrint := hex.EncodeToString(fingerprintBytes)
	return fingerPrint
}

func (kc *KeychainImpl) SignTransaction() string {
	// sign the transcation from here. Receive a transaction payload to this.
	return "signed transaction"
}

func (kc *KeychainImpl) getMasterKey() (*hdkeychain.ExtendedKey, error) {

	// pk := make([]byte, 32)
	// if _, err := rand.Read(pk); err != nil {
	// 	return nil, fmt.Errorf("error generating random bytes: %v", err)
	// }

	pk := getSeed()
	fmt.Println("Size: ", len(pk))

	masterKey, err := hdkeychain.NewMaster(pk, blockchain.GetNetworkParams())
	if err != nil {
		fmt.Println("Error creating master key: ", err)
		return nil, ErrGeneratingKey
	}
	fmt.Println("Master key depth: ", masterKey.Depth())

	return masterKey, nil
}

func getSeed() []byte {

	hexSee := "dc2e7909c0e2e9819c56771351106bd201993843d532ddb4a3ff0faabfd838e1896e1a58ec4b324602cae6138074e88ddd2765e8fb338388cc62d0f282cc838a"
	seed, err := hex.DecodeString(hexSee)
	if err != nil {
		fmt.Println("Error decoding hex seed: ", err)
		return nil
	}
	return seed

}

type AddressInfo struct {
	Fingerprint string
	Xpub        string
}
