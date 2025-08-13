package keychain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/sagarkarki99/db"
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
	mpvk, _ := masterKey.ECPrivKey()
	mPubKey, _ := masterKey.Neuter()
	fmt.Println("Master Key Private Key:", mpvk.Key.String())
	fmt.Println("Master Key Public Key:", mPubKey.String())

	fmt.Println("--------------------------")

	purposeKey, _ := masterKey.Derive(hdkeychain.HardenedKeyStart + 84)
	ppvk, _ := purposeKey.ECPrivKey()
	pPubKey, _ := purposeKey.Neuter()
	fmt.Println("PurposeKey Key private Key:", ppvk.Key.String())
	fmt.Println("PurposeKey Key Public Key:", pPubKey.String())
	fmt.Println("--------------------------")

	// Derive the coin type key for Bitcoin (1 for mainnet, 0 for testnet)
	coinTypeKey, _ := purposeKey.Derive(hdkeychain.HardenedKeyStart + 1)
	ctk, _ := coinTypeKey.ECPrivKey()
	cPubKey, _ := coinTypeKey.Neuter()
	fmt.Println("Cointype Key private Key:", ctk.Key.String())
	fmt.Println("Cointype Key Public Key:", cPubKey.String())
	fmt.Println("--------------------------")

	// Path: m/84'/1'/0'
	// This is the account level. You would use a new index for each account (0, 1, 2...).
	// This is the extended private key you would store for the user.
	// Derive the account key (0 for the first account)
	// This should be incrmental.
	accountKey, err := coinTypeKey.Derive(hdkeychain.HardenedKeyStart + uint32(accountId))
	if err != nil {
		return nil, ErrGeneratingKey
	}

	pvk, _ := accountKey.ECPrivKey()
	xpub, _ := accountKey.Neuter()

	fmt.Println("Account Key Private Key:", pvk.Key.String())
	fmt.Println("Account Key Public Key:", xpub.String())

	pub, _ := accountKey.ECPubKey()

	pubKeyBytes := pub.SerializeCompressed()
	shahash := sha256.Sum256(pubKeyBytes)
	ripeHasher := ripemd160.New()
	ripeHasher.Write(shahash[:])
	hashedRIPEMD160 := ripeHasher.Sum(nil)
	fingerprintBytes := hashedRIPEMD160[:4]

	fingerPrint := hex.EncodeToString(fingerprintBytes)
	fmt.Println("Fingerprint: ", fingerPrint)
	fmt.Println("xPub Key parent fingerprint: ", xpub.ParentFingerprint())

	kc.kr.Save(&db.KeyAddress{
		PrivateKey: pvk.Key.String(),
		PublicKey:  xpub.String(),
	})

	return &AddressInfo{
		Fingerprint: fingerPrint,
		Xpub:        xpub.String(),
	}, nil
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

func (kc *KeychainImpl) getMasterKey() (*hdkeychain.ExtendedKey, error) {

	// pk := make([]byte, 32)
	// if _, err := rand.Read(pk); err != nil {
	// 	return nil, fmt.Errorf("error generating random bytes: %v", err)
	// }

	pk := getSeed()
	fmt.Println("Size: ", len(pk))

	masterKey, err := hdkeychain.NewMaster(pk, &chaincfg.RegressionNetParams)
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
