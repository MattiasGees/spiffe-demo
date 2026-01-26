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

import "errors"

var (
	// ErrAccountNotFound is returned when an account is not found.
	ErrAccountNotFound = errors.New("account not found")

	// ErrInsufficientFunds is returned when an account has insufficient funds for a transfer.
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrInvalidFromAccount is returned when the from account is invalid.
	ErrInvalidFromAccount = errors.New("invalid from account")

	// ErrInvalidToAccount is returned when the to account is invalid.
	ErrInvalidToAccount = errors.New("invalid to account")

	// ErrSameAccount is returned when trying to transfer to the same account.
	ErrSameAccount = errors.New("cannot transfer to the same account")

	// ErrInvalidAmount is returned when the transfer amount is invalid.
	ErrInvalidAmount = errors.New("amount must be greater than zero")

	// ErrTransactionNotFound is returned when a transaction is not found.
	ErrTransactionNotFound = errors.New("transaction not found")
)
