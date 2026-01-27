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
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
)

// RegisterRoutes registers all HTTP routes for the ledger service.
func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	// Account endpoints
	mux.HandleFunc("GET /api/accounts", s.handleGetAccounts)
	mux.HandleFunc("GET /api/accounts/{id}", s.handleGetAccount)

	// Transfer endpoint
	mux.HandleFunc("POST /api/transfers", s.handleCreateTransfer)

	// Transaction endpoints
	mux.HandleFunc("GET /api/transactions", s.handleGetTransactions)
	mux.HandleFunc("GET /api/transactions/{id}", s.handleGetTransaction)

	// Health check
	mux.HandleFunc("GET /health", s.handleHealth)
}

// APIResponse is the standard response wrapper for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// writeSuccess writes a successful JSON response.
func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// writeError writes an error JSON response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error:   message,
	})
}

// logRequest logs the incoming request with SPIFFE ID if available.
func (s *Service) logRequest(r *http.Request, action string) {
	clientID := "unknown"
	if r.TLS != nil {
		if id, err := spiffetls.PeerIDFromConnectionState(*r.TLS); err == nil {
			clientID = id.String()
		}
	}
	log.Printf("[%s] %s %s from %s (SPIFFE: %s)", action, r.Method, r.URL.Path, r.RemoteAddr, clientID)
}

// handleHealth handles GET /health requests.
func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, map[string]string{"status": "healthy"})
}

// handleGetAccounts handles GET /api/accounts requests.
func (s *Service) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "GetAccounts")

	accounts, err := s.store.GetAccounts(r.Context())
	if err != nil {
		log.Printf("Error getting accounts: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to retrieve accounts")
		return
	}

	writeSuccess(w, accounts)
}

// handleGetAccount handles GET /api/accounts/{id} requests.
func (s *Service) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "GetAccount")

	// Parse the account ID from the URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "Account ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid account ID format")
		return
	}

	account, err := s.store.GetAccount(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, "Account not found")
			return
		}
		log.Printf("Error getting account: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}

	writeSuccess(w, account)
}

// handleCreateTransfer handles POST /api/transfers requests.
func (s *Service) handleCreateTransfer(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "CreateTransfer")

	// Validate content type
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	// Parse request body
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create the transfer
	transaction, err := s.store.CreateTransfer(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountNotFound):
			writeError(w, http.StatusNotFound, "Account not found")
		case errors.Is(err, ErrInsufficientFunds):
			writeError(w, http.StatusBadRequest, "Insufficient funds")
		case errors.Is(err, ErrSameAccount):
			writeError(w, http.StatusBadRequest, "Cannot transfer to the same account")
		case errors.Is(err, ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, "Amount must be greater than zero")
		case errors.Is(err, ErrInvalidFromAccount):
			writeError(w, http.StatusBadRequest, "Invalid from account")
		case errors.Is(err, ErrInvalidToAccount):
			writeError(w, http.StatusBadRequest, "Invalid to account")
		default:
			log.Printf("Error creating transfer: %v", err)
			writeError(w, http.StatusInternalServerError, "Failed to create transfer")
		}
		return
	}

	log.Printf("Transfer created: %s -> %s, amount: %s",
		transaction.FromAccount, transaction.ToAccount, transaction.Amount)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    transaction,
	})
}

// handleGetTransactions handles GET /api/transactions requests.
func (s *Service) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "GetTransactions")

	transactions, err := s.store.GetTransactions(r.Context())
	if err != nil {
		log.Printf("Error getting transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	writeSuccess(w, transactions)
}

// handleGetTransaction handles GET /api/transactions/{id} requests.
func (s *Service) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "GetTransaction")

	// Parse the transaction ID from the URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "Transaction ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID format")
		return
	}

	transaction, err := s.store.GetTransaction(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "Transaction not found")
			return
		}
		log.Printf("Error getting transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to retrieve transaction")
		return
	}

	writeSuccess(w, transaction)
}
