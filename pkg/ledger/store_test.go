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
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	checkingID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	savingsID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestMockStore_GetAccounts(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	accounts, err := store.GetAccounts(ctx)
	if err != nil {
		t.Fatalf("GetAccounts() error = %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}
}

func TestMockStore_GetAccount(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name:    "existing account",
			id:      checkingID,
			wantErr: nil,
		},
		{
			name:    "non-existing account",
			id:      uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			wantErr: ErrAccountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := store.GetAccount(ctx, tt.id)
			if err != tt.wantErr {
				t.Errorf("GetAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && account == nil {
				t.Error("GetAccount() returned nil account for existing ID")
			}
		})
	}
}

func TestMockStore_CreateTransfer(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		request TransferRequest
		wantErr error
	}{
		{
			name: "valid transfer",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(100.00),
				Description: "Test transfer",
			},
			wantErr: nil,
		},
		{
			name: "insufficient funds",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(10000.00),
				Description: "Too much",
			},
			wantErr: ErrInsufficientFunds,
		},
		{
			name: "non-existing from account",
			request: TransferRequest{
				FromAccount: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(100.00),
			},
			wantErr: ErrAccountNotFound,
		},
		{
			name: "non-existing to account",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				Amount:      decimal.NewFromFloat(100.00),
			},
			wantErr: ErrAccountNotFound,
		},
		{
			name: "same account",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   checkingID,
				Amount:      decimal.NewFromFloat(100.00),
			},
			wantErr: ErrSameAccount,
		},
		{
			name: "zero amount",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.Zero,
			},
			wantErr: ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh store for each test
			store := NewMockStore()

			transaction, err := store.CreateTransfer(ctx, tt.request)
			if err != tt.wantErr {
				t.Errorf("CreateTransfer() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				if transaction == nil {
					t.Error("CreateTransfer() returned nil transaction for valid request")
				} else {
					if !transaction.Amount.Equal(tt.request.Amount) {
						t.Errorf("Transaction amount = %s, want %s", transaction.Amount, tt.request.Amount)
					}
					if transaction.FromAccount != tt.request.FromAccount {
						t.Errorf("Transaction FromAccount = %s, want %s", transaction.FromAccount, tt.request.FromAccount)
					}
					if transaction.ToAccount != tt.request.ToAccount {
						t.Errorf("Transaction ToAccount = %s, want %s", transaction.ToAccount, tt.request.ToAccount)
					}
				}
			}
		})
	}
}

func TestMockStore_CreateTransfer_UpdatesBalances(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Get initial balances
	checkingBefore, _ := store.GetAccount(ctx, checkingID)
	savingsBefore, _ := store.GetAccount(ctx, savingsID)

	transferAmount := decimal.NewFromFloat(500.00)

	// Perform transfer
	_, err := store.CreateTransfer(ctx, TransferRequest{
		FromAccount: checkingID,
		ToAccount:   savingsID,
		Amount:      transferAmount,
	})
	if err != nil {
		t.Fatalf("CreateTransfer() error = %v", err)
	}

	// Check balances after transfer
	checkingAfter, _ := store.GetAccount(ctx, checkingID)
	savingsAfter, _ := store.GetAccount(ctx, savingsID)

	expectedCheckingBalance := checkingBefore.Balance.Sub(transferAmount)
	expectedSavingsBalance := savingsBefore.Balance.Add(transferAmount)

	if !checkingAfter.Balance.Equal(expectedCheckingBalance) {
		t.Errorf("Checking balance = %s, want %s", checkingAfter.Balance, expectedCheckingBalance)
	}

	if !savingsAfter.Balance.Equal(expectedSavingsBalance) {
		t.Errorf("Savings balance = %s, want %s", savingsAfter.Balance, expectedSavingsBalance)
	}
}

func TestMockStore_GetTransactions(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Initially no transactions
	transactions, err := store.GetTransactions(ctx)
	if err != nil {
		t.Fatalf("GetTransactions() error = %v", err)
	}
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(transactions))
	}

	// Create a transfer
	_, err = store.CreateTransfer(ctx, TransferRequest{
		FromAccount: checkingID,
		ToAccount:   savingsID,
		Amount:      decimal.NewFromFloat(100.00),
	})
	if err != nil {
		t.Fatalf("CreateTransfer() error = %v", err)
	}

	// Should now have 1 transaction
	transactions, err = store.GetTransactions(ctx)
	if err != nil {
		t.Fatalf("GetTransactions() error = %v", err)
	}
	if len(transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(transactions))
	}
}

func TestMockStore_GetTransaction(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Create a transfer
	created, err := store.CreateTransfer(ctx, TransferRequest{
		FromAccount: checkingID,
		ToAccount:   savingsID,
		Amount:      decimal.NewFromFloat(100.00),
	})
	if err != nil {
		t.Fatalf("CreateTransfer() error = %v", err)
	}

	// Get the transaction
	retrieved, err := store.GetTransaction(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTransaction() error = %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("GetTransaction() ID = %s, want %s", retrieved.ID, created.ID)
	}

	// Try to get non-existing transaction
	_, err = store.GetTransaction(ctx, uuid.MustParse("99999999-9999-9999-9999-999999999999"))
	if err != ErrTransactionNotFound {
		t.Errorf("GetTransaction() error = %v, want %v", err, ErrTransactionNotFound)
	}
}

func TestMockStore_Reset(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// Make a transfer
	_, err := store.CreateTransfer(ctx, TransferRequest{
		FromAccount: checkingID,
		ToAccount:   savingsID,
		Amount:      decimal.NewFromFloat(1000.00),
	})
	if err != nil {
		t.Fatalf("CreateTransfer() error = %v", err)
	}

	// Reset the store
	store.Reset()

	// Check balances are reset
	checking, _ := store.GetAccount(ctx, checkingID)
	if !checking.Balance.Equal(decimal.NewFromFloat(5000.00)) {
		t.Errorf("After reset, Checking balance = %s, want 5000.00", checking.Balance)
	}

	// Check transactions are cleared
	transactions, _ := store.GetTransactions(ctx)
	if len(transactions) != 0 {
		t.Errorf("After reset, expected 0 transactions, got %d", len(transactions))
	}
}
