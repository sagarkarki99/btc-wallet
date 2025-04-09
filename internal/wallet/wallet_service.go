package wallet

import (
	"fmt"

	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/keychain"
	repo "github.com/sagarkarki99/internal/repository"
)

type WalletService interface {
	GetDepositAddress(userId string) string
}

func NewWalletService(kc keychain.Keychain) WalletService {
	return &WalletServiceImpl{
		repo: repo.New(),
		kc:   kc,
	}
}

type WalletServiceImpl struct {
	repo repo.WalletRepository
	kc   keychain.Keychain
}

func (ws *WalletServiceImpl) GetDepositAddress(userId string) string {
	wallet, err := ws.repo.Get(userId)
	if wallet != nil {
		return wallet.Address
	}
	addr := ws.kc.GenerateAddress()
	w := db.Wallet{
		Address: addr,
		UserId:  userId,
	}
	ws.repo.Save(w)
	if err != nil {
		fmt.Println("Error getting wallet : ", err)
	}
	return w.Address
}
