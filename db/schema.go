package db

import "time"

// Deprecated: This type is no longer used
type Wallet struct {
	Id string `db:"id"`

	Address string `db:"wallet_address"`
	UserId  string `db:"user_id"`
}

type Address struct {
	Id        int       `db:"id"`
	Hash      string    `db:"address_hash"`
	Index     int       `db:"addr_index"`
	NextIndex int       `db:"next_index"`
	AccountId int       `db:"account_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Account struct {
	Id           int       `db:"id"`
	AccountIndex int       `db:"account_index"`
	XprivKey     string    `db:"xpriv"`
	CreatedAt    time.Time `db:"created_at"`
}

type WalletInfo struct {
	Id               int    `db:"id"`
	XpubKey          string `db:"xpub"`
	AccountId        int    `db:"accountid"`
	NextAddressIndex int    `db:"next_index"`
}
