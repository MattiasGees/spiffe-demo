-- TrustBank database schema
-- This script initializes the database with the required tables and seed data.

-- Create the trustbank database if it doesn't exist
-- Note: This is handled separately in the init script

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

-- Grant permissions to the ledger service user
-- Note: The user is created in the main PostgreSQL init script
GRANT SELECT, INSERT, UPDATE ON accounts TO "spiffe-demo-ledger";
GRANT SELECT, INSERT ON transactions TO "spiffe-demo-ledger";
