CREATE TABLE IF NOT EXISTS account (
    id SERIAL PRIMARY KEY,
    account_index INTEGER NOT NULL,
    xpriv VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS address (
    id SERIAL PRIMARY KEY,
    addr_index INTEGER NOT NULL,
    next_index INTEGER NOT NULL,
    account_id INTEGER REFERENCES account (id),
    address_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallet (
    id SERIAL PRIMARY KEY,
    xpub VARCHAR(255) NOT NULL,
    accountId INTEGER REFERENCES account (id),
    fingerprint VARCHAR(64) NOT NUll
);