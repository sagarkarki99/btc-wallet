package db

type Wallet struct {
	Id string `db:"id"`

	Address string `db:"wallet_address"`
	UserId  string `db:"user_id"`
}

type KeyAddress struct {
	Id         string `db:"id"`
	PrivateKey string `db:"private_key"`
	PublicKey  string `db:"public_key"`
}
