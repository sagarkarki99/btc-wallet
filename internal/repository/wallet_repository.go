package repo

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sagarkarki99/db"
)

type WalletRepository interface {
	Save(w db.Wallet) (string, error)
	Get(userId string) (*db.Wallet, error)
}

func New() WalletRepository {
	return &walletRepository{
		db: db.DB,
	}

}

type walletRepository struct {
	db *sqlx.DB
}

func (db *walletRepository) Save(w db.Wallet) (string, error) {
	var id string
	err := db.db.QueryRow("INSERT INTO wallets (wallet_address, user_id) VALUES ($1, $2) RETURNING id", w.Address, w.UserId).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error while saving wallet: %w", err)
	}
	return id, nil
}

func (d *walletRepository) Get(userId string) (*db.Wallet, error) {
	var w db.Wallet
	err := db.DB.Get(&w, "SELECT * FROM wallets WHERE wallets.user_id = $1", userId)
	if err != nil {
		return nil, fmt.Errorf("error while getting wallet: %w", err)
	}
	return &w, nil
}
