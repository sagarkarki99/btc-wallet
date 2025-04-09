package repo

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sagarkarki99/db"
)

type KeychainRepository interface {
	Save(w *db.KeyAddress) (string, error)
}

func NewKeychainRepo() KeychainRepository {
	return &keychainRepository{
		db: db.DB,
	}
}

type keychainRepository struct {
	db *sqlx.DB
}

func (db *keychainRepository) Save(k *db.KeyAddress) (string, error) {
	var id string
	err := db.db.QueryRow("INSERT INTO key_addresses (private_key, public_key) VALUES ($1, $2) RETURNING id", k.PrivateKey, k.PublicKey).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error while saving wallet: %w", err)
	}
	return id, nil
}
