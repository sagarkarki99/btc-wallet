package wallet

import (
	"testing"

	"github.com/sagarkarki99/db"
	"github.com/sagarkarki99/internal/keychain"
	repo "github.com/sagarkarki99/internal/repository"
)

type MockRepo struct {
	SaveAddressFunc func(w db.Address) (string, error)
	SaveAccountFunc func(w db.Account) (int, error)
	GetAccountFunc  func(userId int) (*db.Account, error)
	GetAddressFunc  func(userId int) (*db.Address, error)
}

func (m MockRepo) SaveAddress(w db.Address) (string, error) {
	return "", nil
}

func (m MockRepo) SaveAccount(w db.Account) (int, error) {
	return 0, nil
}

func (m MockRepo) GetAccount(userId int) (*db.Account, error) {
	return nil, nil
}

func (m MockRepo) GetAddress(userId int) (*db.Address, error) {
	return nil, nil
}

type MockKeychain struct {
}

func (k MockKeychain) GenerateAddress(accountId uint32) (*keychain.AddressInfo, error) {
	return nil, nil
}

func (k MockKeychain) SignTransaction() string {
	return ""
}

func NewWalletServiceTest(repo repo.WalletRepository, kc keychain.Keychain) WalletService {
	return &WalletServiceImpl{
		repo: repo,
		kc:   kc,
	}
}

func TestAddingUtxo(t *testing.T) {
	//arrange
	mockRepo :=
		&MockRepo{
			SaveAddressFunc: func(w db.Address) (string, error) {
				return "", nil
			},
		}

	mockKc := &MockKeychain{}
	NewWalletServiceTest(mockRepo, mockKc)

	// act

	//assert
}
