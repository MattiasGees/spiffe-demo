/*
Copyright © 2024 Mattias Gees mattias.gees@venafi.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ledger

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// PostgresStore implements the Store interface using PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// PostgresConfig contains the configuration for connecting to PostgreSQL.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Database string
	// SSLMode specifies the SSL mode (disable, require, verify-ca, verify-full)
	SSLMode string
	// SSLCert is the path to the client certificate file (for X.509 auth)
	SSLCert string
	// SSLKey is the path to the client private key file (for X.509 auth)
	SSLKey string
	// SSLRootCert is the path to the root CA certificate file
	SSLRootCert string
}

// NewPostgresStore creates a new PostgresStore with the given configuration.
func NewPostgresStore(ctx context.Context, cfg PostgresConfig) (*PostgresStore, error) {
	connString := buildConnectionString(cfg)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure TLS if SSL certificates are provided
	if cfg.SSLCert != "" && cfg.SSLKey != "" {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		poolConfig.ConnConfig.TLSConfig = tlsConfig
	}

	// Set pool configuration
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// buildConnectionString builds a PostgreSQL connection string from the config.
func buildConnectionString(cfg PostgresConfig) string {
	port := cfg.Port
	if port == 0 {
		port = 5432
	}

	database := cfg.Database
	if database == "" {
		database = "trustbank"
	}

	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		cfg.Host, port, cfg.User, database, sslMode,
	)
}

// buildTLSConfig builds a TLS configuration for X.509 certificate authentication.
func buildTLSConfig(cfg PostgresConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.SSLCert, cfg.SSLKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if cfg.SSLRootCert != "" {
		rootCert, err := os.ReadFile(cfg.SSLRootCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read root certificate: %w", err)
		}

		rootCertPool := x509.NewCertPool()
		if !rootCertPool.AppendCertsFromPEM(rootCert) {
			return nil, fmt.Errorf("failed to append root certificate")
		}
		tlsConfig.RootCAs = rootCertPool
	}

	return tlsConfig, nil
}

// Close closes the database connection pool.
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

// GetAccounts returns all accounts.
func (s *PostgresStore) GetAccounts(ctx context.Context) ([]Account, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, balance, created_at
		FROM accounts
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		var balanceStr string
		if err := rows.Scan(&a.ID, &a.Name, &balanceStr, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		a.Balance, err = decimal.NewFromString(balanceStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse balance: %w", err)
		}
		accounts = append(accounts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}

// GetAccount returns a single account by ID.
func (s *PostgresStore) GetAccount(ctx context.Context, id uuid.UUID) (*Account, error) {
	var a Account
	var balanceStr string

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, balance, created_at
		FROM accounts
		WHERE id = $1
	`, id).Scan(&a.ID, &a.Name, &balanceStr, &a.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	a.Balance, err = decimal.NewFromString(balanceStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	return &a, nil
}

// GetTransactions returns all transactions, ordered by creation time (newest first).
func (s *PostgresStore) GetTransactions(ctx context.Context) ([]Transaction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, from_account, to_account, amount, description, created_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var amountStr string
		var description *string
		if err := rows.Scan(&t.ID, &t.FromAccount, &t.ToAccount, &amountStr, &description, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		t.Amount, err = decimal.NewFromString(amountStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}
		if description != nil {
			t.Description = *description
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// GetTransaction returns a single transaction by ID.
func (s *PostgresStore) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	var t Transaction
	var amountStr string
	var description *string

	err := s.pool.QueryRow(ctx, `
		SELECT id, from_account, to_account, amount, description, created_at
		FROM transactions
		WHERE id = $1
	`, id).Scan(&t.ID, &t.FromAccount, &t.ToAccount, &amountStr, &description, &t.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrTransactionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction: %w", err)
	}

	t.Amount, err = decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}
	if description != nil {
		t.Description = *description
	}

	return &t, nil
}

// CreateTransfer creates a new transfer between accounts.
// This operation is atomic - either both accounts are updated or neither is.
func (s *PostgresStore) CreateTransfer(ctx context.Context, req TransferRequest) (*Transaction, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Start a transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock and check the from account
	var fromBalance string
	err = tx.QueryRow(ctx, `
		SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
	`, req.FromAccount).Scan(&fromBalance)
	if err == pgx.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lock from account: %w", err)
	}

	fromBalanceDec, err := decimal.NewFromString(fromBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to parse from balance: %w", err)
	}

	if fromBalanceDec.LessThan(req.Amount) {
		return nil, ErrInsufficientFunds
	}

	// Lock and check the to account exists
	var toBalance string
	err = tx.QueryRow(ctx, `
		SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
	`, req.ToAccount).Scan(&toBalance)
	if err == pgx.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lock to account: %w", err)
	}

	// Deduct from the from account
	_, err = tx.Exec(ctx, `
		UPDATE accounts SET balance = balance - $1 WHERE id = $2
	`, req.Amount.String(), req.FromAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct from account: %w", err)
	}

	// Add to the to account
	_, err = tx.Exec(ctx, `
		UPDATE accounts SET balance = balance + $1 WHERE id = $2
	`, req.Amount.String(), req.ToAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to add to account: %w", err)
	}

	// Create the transaction record
	transaction := Transaction{
		ID:          uuid.New(),
		FromAccount: req.FromAccount,
		ToAccount:   req.ToAccount,
		Amount:      req.Amount,
		Description: req.Description,
		CreatedAt:   time.Now().UTC(),
	}

	var descPtr *string
	if req.Description != "" {
		descPtr = &req.Description
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (id, from_account, to_account, amount, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, transaction.ID, transaction.FromAccount, transaction.ToAccount, transaction.Amount.String(), descPtr, transaction.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, nil
}
