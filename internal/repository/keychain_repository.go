package repo

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sagarkarki99/db"
)

type KeychainRepository interface {
	Save(w *db.Account) (int, error)
	GetLatestAccountIndex() (*db.Account, error)
}

func NewKeychainRepo() KeychainRepository {
	return &keychainRepository{
		db: db.DB,
	}
}

type keychainRepository struct {
	db *sqlx.DB
}

func (db *keychainRepository) Save(a *db.Account) (int, error) {

	var id int
	err := db.db.QueryRow("INSERT INTO account (xpriv,account_index) VALUES ($1,$2) RETURNING id", a.XprivKey, a.AccountIndex).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error while saving Account: %w", err)
	}
	return id, nil
}

func (d *keychainRepository) GetLatestAccountIndex() (*db.Account, error) {
	var w db.Account
	err := d.db.Get(&w, "SELECT * FROM account ORDER BY account.account_index LIMIT 1 ")
	if err != nil && strings.Contains(err.Error(), "sql: no rows in result set") {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error while getting account: %w", err)
	}
	return &w, nil
}
