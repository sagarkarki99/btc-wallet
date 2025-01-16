package repo

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sagarkarki99/db"
)

type WalletRepository interface {
	Save(w Wallet) (string, error)
	Get(walletId string) (*Wallet, error)
}

type Wallet struct {
	PrivateKey string
	PublicKey  string
	UserId     string
}

func New() *walletRepository {
	return &walletRepository{
		db: db.DB,
	}

}

type walletRepository struct {
	db *sqlx.DB
}

func (db *walletRepository) Save(w Wallet) (string, error) {
	var id string
	err := db.db.QueryRow("INSERT INTO wallet (private_key, public_key, user_id) VALUES ($1, $2, $3) RETURNING id", w.PrivateKey, w.PublicKey, w.UserId).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error while saving wallet: %w", err)
	}
	return id, nil
}

func (db *walletRepository) Get(walletId string) (*Wallet, error) {
	var w Wallet
	err := db.db.Get(&w, "SELECT * FROM wallet WHERE id = $1", walletId)
	if err != nil {
		return nil, fmt.Errorf("error while getting wallet: %w", err)
	}
	return &w, nil
}
