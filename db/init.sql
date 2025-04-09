CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    wallet_address VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL
);


CREATE TABLE IF NOT EXISTS key_addresses (
    id SERIAL PRIMARY KEY,
    private_key VARCHAR(255) NOT NULL,
    public_key VARCHAR(255) NOT NULL
);
