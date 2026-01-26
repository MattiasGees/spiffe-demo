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

// Package ledger provides the TrustBank ledger service for managing accounts and transactions.
package ledger

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Account represents a bank account in the TrustBank system.
type Account struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
}

// Transaction represents a money transfer between accounts.
type Transaction struct {
	ID          uuid.UUID       `json:"id"`
	FromAccount uuid.UUID       `json:"from_account"`
	ToAccount   uuid.UUID       `json:"to_account"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TransferRequest represents a request to transfer money between accounts.
type TransferRequest struct {
	FromAccount uuid.UUID       `json:"from_account"`
	ToAccount   uuid.UUID       `json:"to_account"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description,omitempty"`
}

// Validate checks if the transfer request is valid.
func (t *TransferRequest) Validate() error {
	if t.FromAccount == uuid.Nil {
		return ErrInvalidFromAccount
	}
	if t.ToAccount == uuid.Nil {
		return ErrInvalidToAccount
	}
	if t.FromAccount == t.ToAccount {
		return ErrSameAccount
	}
	if t.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	return nil
}

// AccountsResponse represents the response for listing accounts.
type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

// TransactionsResponse represents the response for listing transactions.
type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

// TransactionResponse represents the response for a single transaction.
type TransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
