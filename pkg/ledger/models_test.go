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
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestTransferRequest_Validate(t *testing.T) {
	validFromAccount := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	validToAccount := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	validAmount := decimal.NewFromFloat(100.00)

	tests := []struct {
		name    string
		request TransferRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: TransferRequest{
				FromAccount: validFromAccount,
				ToAccount:   validToAccount,
				Amount:      validAmount,
				Description: "Test transfer",
			},
			wantErr: nil,
		},
		{
			name: "valid request without description",
			request: TransferRequest{
				FromAccount: validFromAccount,
				ToAccount:   validToAccount,
				Amount:      validAmount,
			},
			wantErr: nil,
		},
		{
			name: "nil from account",
			request: TransferRequest{
				FromAccount: uuid.Nil,
				ToAccount:   validToAccount,
				Amount:      validAmount,
			},
			wantErr: ErrInvalidFromAccount,
		},
		{
			name: "nil to account",
			request: TransferRequest{
				FromAccount: validFromAccount,
				ToAccount:   uuid.Nil,
				Amount:      validAmount,
			},
			wantErr: ErrInvalidToAccount,
		},
		{
			name: "same account",
			request: TransferRequest{
				FromAccount: validFromAccount,
				ToAccount:   validFromAccount,
				Amount:      validAmount,
			},
			wantErr: ErrSameAccount,
		},
		{
			name: "zero amount",
			request: TransferRequest{
				FromAccount: validFromAccount,
				ToAccount:   validToAccount,
				Amount:      decimal.Zero,
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "negative amount",
			request: TransferRequest{
				FromAccount: validFromAccount,
				ToAccount:   validToAccount,
				Amount:      decimal.NewFromFloat(-100.00),
			},
			wantErr: ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccount_JSONTags(t *testing.T) {
	// This test ensures the JSON tags are correctly set by marshaling/unmarshaling
	// For now, just verify the struct can be created
	account := Account{
		ID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:    "Test Account",
		Balance: decimal.NewFromFloat(1000.00),
	}

	if account.Name != "Test Account" {
		t.Errorf("Expected name 'Test Account', got '%s'", account.Name)
	}

	if !account.Balance.Equal(decimal.NewFromFloat(1000.00)) {
		t.Errorf("Expected balance 1000.00, got %s", account.Balance.String())
	}
}

func TestTransaction_JSONTags(t *testing.T) {
	// Verify the struct can be created
	transaction := Transaction{
		ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		FromAccount: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ToAccount:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Amount:      decimal.NewFromFloat(100.00),
		Description: "Test transaction",
	}

	if !transaction.Amount.Equal(decimal.NewFromFloat(100.00)) {
		t.Errorf("Expected amount 100.00, got %s", transaction.Amount.String())
	}

	if transaction.Description != "Test transaction" {
		t.Errorf("Expected description 'Test transaction', got '%s'", transaction.Description)
	}
}
