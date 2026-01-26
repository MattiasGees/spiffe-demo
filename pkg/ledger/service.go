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
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// Service represents the ledger service with its dependencies.
type Service struct {
	store         Store
	serverAddress string
	spiffeAuthz   string
}

// ServiceConfig contains the configuration for the ledger service.
type ServiceConfig struct {
	// ServerAddress is the address to bind the server to (e.g., "0.0.0.0:8080")
	ServerAddress string
	// SpiffeAuthz is the authorized SPIFFE ID that can connect to this service
	SpiffeAuthz string
	// Store is the data store implementation
	Store Store
}

// NewService creates a new ledger service with the given configuration.
func NewService(cfg ServiceConfig) *Service {
	return &Service{
		store:         cfg.Store,
		serverAddress: cfg.ServerAddress,
		spiffeAuthz:   cfg.SpiffeAuthz,
	}
}

// Run starts the ledger service with SPIFFE mTLS authentication.
//
// SPIFFE CONCEPT: SPIFFE-Native Banking Service
// This ledger service demonstrates how to build a secure banking API that natively
// understands SPIFFE. The service uses the SPIFFE Workload API to automatically
// obtain and rotate X.509 certificates (SVIDs). Only clients with the authorized
// SPIFFE ID can connect - implementing zero-trust identity-based access control.
func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up HTTP routes
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	// SPIFFE CONCEPT: Server-Side X509Source
	// The X509Source connects to the SPIFFE Workload API (typically via a Unix socket)
	// to obtain the server's X.509-SVID (SPIFFE Verifiable Identity Document).
	// This certificate will be presented to clients during TLS handshake.
	// Certificate rotation is automatic - SPIRE handles renewal before expiry.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return fmt.Errorf("unable to create X509Source: %w", err)
	}
	defer source.Close()

	// SPIFFE CONCEPT: Client Authorization
	// Parse the authorized SPIFFE ID that is allowed to connect to this service.
	// In the TrustBank demo, this would typically be the customer service's SPIFFE ID.
	// Only workloads with this exact SPIFFE ID will be accepted.
	clientID, err := spiffeid.FromString(s.spiffeAuthz)
	if err != nil {
		return fmt.Errorf("invalid SPIFFE ID configuration: %w", err)
	}

	log.Printf("Starting ledger service on %s", s.serverAddress)
	log.Printf("Authorized SPIFFE ID: %s", clientID.String())

	// SPIFFE CONCEPT: mTLS Server Configuration
	// MTLSServerConfig creates a TLS configuration for mutual TLS on the server side:
	//   - First 'source' parameter: provides server certificate to present to clients
	//   - Second 'source' parameter: provides trust bundles to validate client certificates
	//   - AuthorizeID(clientID): only accept clients with this exact SPIFFE ID
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))
	server := &http.Server{
		Addr:              s.serverAddress,
		Handler:           mux,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: time.Second * 10,
	}

	// Start the SPIFFE mTLS server.
	// Empty strings for cert/key files because TLSConfig already contains our SVID.
	if err := server.ListenAndServeTLS("", ""); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// StartServer is the entry point called from the CLI command.
func StartServer(spiffeAuthz, serverAddress string, store Store) {
	service := NewService(ServiceConfig{
		ServerAddress: serverAddress,
		SpiffeAuthz:   spiffeAuthz,
		Store:         store,
	})

	if err := service.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
