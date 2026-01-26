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

	"github.com/google/uuid"
)

// Store defines the interface for ledger data persistence.
// This interface allows for easy mocking in tests.
type Store interface {
	// GetAccounts returns all accounts.
	GetAccounts(ctx context.Context) ([]Account, error)

	// GetAccount returns a single account by ID.
	GetAccount(ctx context.Context, id uuid.UUID) (*Account, error)

	// GetTransactions returns all transactions, ordered by creation time (newest first).
	GetTransactions(ctx context.Context) ([]Transaction, error)

	// GetTransaction returns a single transaction by ID.
	GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error)

	// CreateTransfer creates a new transfer between accounts.
	// This operation is atomic - either both accounts are updated or neither is.
	CreateTransfer(ctx context.Context, req TransferRequest) (*Transaction, error)

	// Close closes the store connection.
	Close() error
}
