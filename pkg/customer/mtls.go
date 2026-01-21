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
package customer

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// Handles requests for connecting to the SPIFFE native backend
func (c *CustomerService) mtlsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the rootHandler from %s", r.RemoteAddr)
	mTLSCall(w, c.spiffeAuthz, c.backendService)
}

// General mTLS call to SPIFFE enabled servers. This can be either a SPIFFE native application or a webserver/apiserver that is fronted by a SPIFFE proxy like Envoy.
//
// SPIFFE CONCEPT: Application-Layer mTLS
// This demonstrates how an application can use SPIFFE identities directly in code
// to establish mutually authenticated TLS connections. The go-spiffe library handles
// all certificate management automatically - no need to deal with certificate files,
// rotation, or manual TLS configuration.
func mTLSCall(w http.ResponseWriter, spiffeAuthZ string, backendAddress string) {
	w.Header().Set("Content-Type", "text/html")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// SPIFFE CONCEPT: X509Source and the Workload API
	// The X509Source automatically connects to the SPIFFE Workload API (typically provided
	// by SPIRE Agent) via a Unix domain socket. It fetches the workload's X.509-SVID
	// (SPIFFE Verifiable Identity Document) which contains:
	//   - The workload's SPIFFE ID in the certificate's URI SAN (e.g., spiffe://example.org/myservice)
	//   - A private key for proving identity
	//   - Trust bundles for validating other workloads' certificates
	// The source automatically handles certificate rotation - when SPIRE rotates the SVID,
	// the source gets updated transparently.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create X509Source: %v", err), http.StatusInternalServerError)
		return
	}
	defer source.Close()

	// SPIFFE CONCEPT: SPIFFE ID Authorization
	// A SPIFFE ID is a URI that uniquely identifies a workload (e.g., spiffe://trust-domain/path).
	// Here we parse the expected server's SPIFFE ID that we want to connect to.
	// This implements zero-trust networking: we don't just accept "any valid certificate",
	// we verify the exact identity we expect to communicate with.
	serverID, err := spiffeid.FromString(spiffeAuthZ)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid SPIFFE ID configuration: %v", err), http.StatusInternalServerError)
		return
	}

	// SPIFFE CONCEPT: mTLS Client Configuration
	// MTLSClientConfig creates a TLS configuration for mutual TLS:
	//   - First 'source' parameter: provides our client certificate (X.509-SVID) to present to the server
	//   - Second 'source' parameter: provides trust bundles to validate the server's certificate
	//   - AuthorizeID(serverID): only accept connections to servers with this exact SPIFFE ID
	// This ensures both sides prove their identity - true mutual authentication.
	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeID(serverID))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Do a GET call to the backend and get the response.
	resp, err := client.Get(backendAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to %q: %v", backendAddress, err), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	// Read the body from the response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read body: %v", err), http.StatusInternalServerError)
		return
	}

	// SPIFFE CONCEPT: Retrieving Peer Identity
	// After establishing the mTLS connection, we can extract the server's SPIFFE ID
	// from the TLS connection state. This is useful for logging, auditing, or making
	// authorization decisions based on the authenticated identity.
	serverSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*resp.TLS)
	if err != nil {
		http.Error(w, fmt.Sprintf("Wasn't able to determine the SPIFFE ID of the server: %v", err), http.StatusInternalServerError)
		return
	}

	// Showcase the retrieved information and send it back to the customer.
	fmt.Fprintf(w, "<p>Got a response from: %s</p>", serverSPIFFEID.String())
	fmt.Fprintf(w, "<p>Server says: %q</p>", body)
}
