package repo

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sagarkarki99/db"
)

var (
	ErrNotFound = errors.New("not found")
)

type WalletRepository interface {
	SaveAddress(w db.Address) (string, error)
	GetAddress(userId int) (*db.Address, error)
	GetWalletInfo(userId int) (*db.WalletInfo, error)
	SaveWallet(xpub string, accountId int, fingerprint string) (int, error)
}

func New() WalletRepository {
	return &walletRepository{
		db: db.DB,
	}

}

type walletRepository struct {
	db *sqlx.DB
}

func (d *walletRepository) GetWalletInfo(userId int) (*db.WalletInfo, error) {
	wallet := &db.WalletInfo{
		Id:               1,
		XpubKey:          "",
		NextAddressIndex: 1,
		AccountId:        1,
	}
	return wallet, nil
}

func (db *walletRepository) SaveAddress(w db.Address) (string, error) {
	var id string
	err := db.db.QueryRow("INSERT INTO address (addr_index, account_id, address_hash,next_index) VALUES ($1, $2,$3,$4) RETURNING id", w.Index, w.AccountId, w.Hash, w.NextIndex).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error while saving address: %w", err)
	}
	return id, nil
}

func (db *walletRepository) SaveWallet(xpub string, accountId int, fingerprint string) (int, error) {
	var id int
	err := db.db.QueryRow("INSERT INTO wallet (xpub, accountid, fingerprint) VALUES ($1, $2, $3) RETURNING id", xpub, accountId, fingerprint).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error while saving address: %w", err)
	}
	return id, nil
}

func (d *walletRepository) GetAddress(accountId int) (*db.Address, error) {
	var w db.Address
	err := db.DB.Get(&w, "SELECT * FROM address WHERE address.account_id = $1 ORDER BY next_index DESC LIMIT 1", accountId)
	if err != nil {
		return nil, fmt.Errorf("error while getting wallet: %w", err)
	}
	return &w, nil
}
