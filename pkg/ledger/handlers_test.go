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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func newTestService() *Service {
	return NewService(ServiceConfig{
		Store: NewMockStore(),
	})
}

func TestHandleHealth(t *testing.T) {
	service := newTestService()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	service.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestHandleGetAccounts(t *testing.T) {
	service := newTestService()

	req := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	w := httptest.NewRecorder()

	service.handleGetAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	// Check that we got accounts
	accounts, ok := response.Data.([]interface{})
	if !ok {
		t.Fatal("Expected Data to be an array")
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}
}

func TestHandleGetAccount(t *testing.T) {
	service := newTestService()

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "valid account",
			id:             "11111111-1111-1111-1111-111111111111",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "non-existing account",
			id:             "99999999-9999-9999-9999-999999999999",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:           "invalid UUID format",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/accounts/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			service.handleGetAccount(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response APIResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Success != tt.expectSuccess {
				t.Errorf("Expected success=%v, got %v", tt.expectSuccess, response.Success)
			}
		})
	}
}

func TestHandleCreateTransfer(t *testing.T) {
	checkingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	savingsID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	tests := []struct {
		name           string
		request        TransferRequest
		contentType    string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "valid transfer",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(100.00),
				Description: "Test transfer",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "insufficient funds",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(10000.00),
				Description: "Too much",
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "same account",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   checkingID,
				Amount:      decimal.NewFromFloat(100.00),
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "zero amount",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.Zero,
			},
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "non-existing from account",
			request: TransferRequest{
				FromAccount: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(100.00),
			},
			contentType:    "application/json",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name: "wrong content type",
			request: TransferRequest{
				FromAccount: checkingID,
				ToAccount:   savingsID,
				Amount:      decimal.NewFromFloat(100.00),
			},
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh service for each test
			service := newTestService()

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/transfers", bytes.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			service.handleCreateTransfer(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response APIResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Success != tt.expectSuccess {
				t.Errorf("Expected success=%v, got %v", tt.expectSuccess, response.Success)
			}
		})
	}
}

func TestHandleGetTransactions(t *testing.T) {
	service := newTestService()

	// Initially no transactions
	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	w := httptest.NewRecorder()

	service.handleGetTransactions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestHandleGetTransaction(t *testing.T) {
	service := newTestService()

	// First create a transaction
	checkingID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	savingsID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	body, _ := json.Marshal(TransferRequest{
		FromAccount: checkingID,
		ToAccount:   savingsID,
		Amount:      decimal.NewFromFloat(100.00),
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/transfers", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	service.handleCreateTransfer(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Failed to create transfer: %d", createW.Code)
	}

	var createResponse APIResponse
	if err := json.NewDecoder(createW.Body).Decode(&createResponse); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	// Extract transaction ID
	transactionData, ok := createResponse.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected Data to be a map")
	}
	transactionID := transactionData["id"].(string)

	// Test getting the transaction
	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "existing transaction",
			id:             transactionID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "non-existing transaction",
			id:             "99999999-9999-9999-9999-999999999999",
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
		},
		{
			name:           "invalid UUID format",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/transactions/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			service.handleGetTransaction(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response APIResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Success != tt.expectSuccess {
				t.Errorf("Expected success=%v, got %v", tt.expectSuccess, response.Success)
			}
		})
	}
}

func TestHandleCreateTransfer_InvalidJSON(t *testing.T) {
	service := newTestService()

	req := httptest.NewRequest(http.MethodPost, "/api/transfers", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	service.handleCreateTransfer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("Expected success to be false")
	}
}
