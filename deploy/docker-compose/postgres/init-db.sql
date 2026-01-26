-- TrustBank database initialization for Docker Compose testing
-- This script runs after PostgreSQL starts

-- Create the ledger user (matches the SPIFFE ID path component)
-- The CN in the X.509-SVID will be "ledger" based on spiffe://example.org/ledger
CREATE USER ledger;

-- Grant permissions to the ledger user
GRANT CONNECT ON DATABASE trustbank TO ledger;

-- Create tables in trustbank database
\c trustbank

-- Accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT balance_non_negative CHECK (balance >= 0)
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    from_account UUID NOT NULL REFERENCES accounts(id),
    to_account UUID NOT NULL REFERENCES accounts(id),
    amount DECIMAL(15,2) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT amount_positive CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (from_account != to_account)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);

-- Seed data: Create demo accounts
INSERT INTO accounts (id, name, balance, created_at) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Checking', 5000.00, NOW()),
    ('22222222-2222-2222-2222-222222222222', 'Savings', 5000.00, NOW())
ON CONFLICT (id) DO NOTHING;

-- Grant table permissions to ledger user
GRANT SELECT, INSERT, UPDATE ON accounts TO ledger;
GRANT SELECT, INSERT ON transactions TO ledger;

-- Show confirmation
\echo 'TrustBank database initialized successfully'
\echo 'User "ledger" created with certificate authentication'
