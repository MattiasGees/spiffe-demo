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
package backend

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mattiasgees/spiffe-demo/pkg/common"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type BackendService struct {
	spiffeAuthz   string
	serverAddress string
}

// Main function that creates the backend server and starts it. This is called from the CLI.
func StartServer(spiffeAuthz, serverAddress string) {
	backendService := BackendService{
		spiffeAuthz:   spiffeAuthz,
		serverAddress: serverAddress,
	}

	if err := backendService.run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// This gets called from the main function and actually starts an mTLS server that is SPIFFE capable.
//
// SPIFFE CONCEPT: SPIFFE-Native Server
// This backend service demonstrates how to build a server that natively understands SPIFFE.
// Instead of configuring TLS with certificate files, we use the SPIFFE Workload API to
// automatically obtain and rotate certificates. The server only accepts connections from
// clients with specific SPIFFE IDs - implementing identity-based access control.
func (b *BackendService) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up a `/` resource handler
	http.HandleFunc("/", b.rootHandler)

	// SPIFFE CONCEPT: Server-Side X509Source
	// Just like the client, the server uses X509Source to get its identity from SPIRE.
	// The server's X.509-SVID will be presented to clients during the TLS handshake,
	// allowing clients to verify they're talking to the right service.
	// Certificate rotation is automatic - SPIRE handles renewal before expiry.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return fmt.Errorf("unable to create X509Source: %w", err)
	}
	defer source.Close()

	// SPIFFE CONCEPT: Client Authorization
	// The server specifies which SPIFFE ID(s) are allowed to connect.
	// This is the "authorization" part of authentication + authorization.
	// Only workloads with this exact SPIFFE ID will be accepted.
	// You can also use AuthorizeAny() to accept any valid SPIFFE ID,
	// or implement custom authorization logic with AuthorizeFunc().
	clientID, err := spiffeid.FromString(b.spiffeAuthz)
	if err != nil {
		return fmt.Errorf("invalid SPIFFE ID configuration: %w", err)
	}

	// SPIFFE CONCEPT: mTLS Server Configuration
	// MTLSServerConfig creates a TLS configuration for mutual TLS on the server side:
	//   - First 'source' parameter: provides server certificate to present to clients
	//   - Second 'source' parameter: provides trust bundles to validate client certificates
	//   - AuthorizeID(clientID): only accept clients with this exact SPIFFE ID
	// Note: ListenAndServeTLS("", "") works because the TLS config already has the certs!
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))
	server := &http.Server{
		Addr:              b.serverAddress,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: time.Second * 10,
	}

	// Serve the SPIFFE mTLS server.
	// Empty strings for cert/key files because TLSConfig already contains our SVID.
	if err := server.ListenAndServeTLS("", ""); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// function that handles calls to `/`. This will just respond with a simple message and the date and time.
func (b *BackendService) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received from %s", r.RemoteAddr)
	currentTime := time.Now()
	formattedTime := currentTime.Format(common.TimeFormat)
	text := fmt.Sprintf("%s: Successfully connected to the backend service!!!", formattedTime)
	if _, err := io.WriteString(w, text); err != nil {
		log.Printf("Error writing response: %v", err)
	}

	// SPIFFE CONCEPT: Identifying the Caller
	// In a SPIFFE-authenticated connection, we can extract the client's SPIFFE ID
	// from the TLS connection state. This enables identity-aware logging and
	// fine-grained authorization decisions within the request handler.
	requestorSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*r.TLS)
	if err != nil {
		log.Printf("Wasn't able to determine the SPIFFE ID of the requestor: %v", err)
	}
	log.Printf("Responded to %s with the following message: %s", requestorSPIFFEID.String(), text)
}
