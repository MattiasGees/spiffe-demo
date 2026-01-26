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
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MockStore is an in-memory implementation of the Store interface for testing.
type MockStore struct {
	mu           sync.RWMutex
	accounts     map[uuid.UUID]*Account
	transactions []Transaction
}

// NewMockStore creates a new MockStore with default test accounts.
func NewMockStore() *MockStore {
	store := &MockStore{
		accounts:     make(map[uuid.UUID]*Account),
		transactions: make([]Transaction, 0),
	}

	// Add default test accounts
	checkingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	savingsID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	store.accounts[checkingID] = &Account{
		ID:        checkingID,
		Name:      "Checking",
		Balance:   decimal.NewFromFloat(5000.00),
		CreatedAt: time.Now().UTC(),
	}

	store.accounts[savingsID] = &Account{
		ID:        savingsID,
		Name:      "Savings",
		Balance:   decimal.NewFromFloat(5000.00),
		CreatedAt: time.Now().UTC(),
	}

	return store
}

// GetAccounts returns all accounts.
func (s *MockStore) GetAccounts(ctx context.Context) ([]Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		accounts = append(accounts, *a)
	}
	return accounts, nil
}

// GetAccount returns a single account by ID.
func (s *MockStore) GetAccount(ctx context.Context, id uuid.UUID) (*Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	// Return a copy to prevent external modification
	accountCopy := *account
	return &accountCopy, nil
}

// GetTransactions returns all transactions, ordered by creation time (newest first).
func (s *MockStore) GetTransactions(ctx context.Context) ([]Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy of the transactions slice
	transactions := make([]Transaction, len(s.transactions))
	copy(transactions, s.transactions)
	return transactions, nil
}

// GetTransaction returns a single transaction by ID.
func (s *MockStore) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.transactions {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, ErrTransactionNotFound
}

// CreateTransfer creates a new transfer between accounts.
func (s *MockStore) CreateTransfer(ctx context.Context, req TransferRequest) (*Transaction, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check from account exists and has sufficient funds
	fromAccount, ok := s.accounts[req.FromAccount]
	if !ok {
		return nil, ErrAccountNotFound
	}

	if fromAccount.Balance.LessThan(req.Amount) {
		return nil, ErrInsufficientFunds
	}

	// Check to account exists
	toAccount, ok := s.accounts[req.ToAccount]
	if !ok {
		return nil, ErrAccountNotFound
	}

	// Perform the transfer
	fromAccount.Balance = fromAccount.Balance.Sub(req.Amount)
	toAccount.Balance = toAccount.Balance.Add(req.Amount)

	// Create transaction record
	transaction := Transaction{
		ID:          uuid.New(),
		FromAccount: req.FromAccount,
		ToAccount:   req.ToAccount,
		Amount:      req.Amount,
		Description: req.Description,
		CreatedAt:   time.Now().UTC(),
	}

	// Prepend to transactions list (newest first)
	s.transactions = append([]Transaction{transaction}, s.transactions...)

	return &transaction, nil
}

// Close closes the store (no-op for mock).
func (s *MockStore) Close() error {
	return nil
}

// Reset resets the mock store to its initial state (useful for tests).
func (s *MockStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	checkingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	savingsID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	s.accounts[checkingID].Balance = decimal.NewFromFloat(5000.00)
	s.accounts[savingsID].Balance = decimal.NewFromFloat(5000.00)
	s.transactions = make([]Transaction, 0)
}
