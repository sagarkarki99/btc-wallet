package repo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sagarkarki99/db"
)

var (
	ErrNotFound = errors.New("not found")
)

type WalletRepository interface {
	SaveAddress(w db.Address) (string, error)
	GetAddress(userId int) (*db.Address, error)
	SaveAccount(w db.Account) (int, error)
	GetAccount(userId int) (*db.Account, error)
}

func New() WalletRepository {
	return &walletRepository{
		db: db.DB,
	}

}

type walletRepository struct {
	db *sqlx.DB
}

func (db *walletRepository) SaveAccount(w db.Account) (int, error) {
	var id int
	err := db.db.QueryRow("INSERT INTO account (xpub,fingerprint) VALUES ($1,$2) RETURNING id", w.XpubKey, w.Fingerprint).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error while saving Account: %w", err)
	}
	return id, nil
}

func (d *walletRepository) GetAccount(userId int) (*db.Account, error) {
	var w db.Account
	err := db.DB.Get(&w, "SELECT * FROM account WHERE account.id = $1", userId)
	if err != nil && strings.Contains(err.Error(), "sql: no rows in result set") {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error while getting account: %w", err)
	}
	return &w, nil
}

func (db *walletRepository) SaveAddress(w db.Address) (string, error) {
	var id string
	err := db.db.QueryRow("INSERT INTO address (addr_index, account_id, address_hash,next_index) VALUES ($1, $2,$3,$4) RETURNING id", w.Index, w.AccountId, w.Hash, w.NextIndex).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error while saving address: %w", err)
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
