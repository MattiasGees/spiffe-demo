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
func mTLSCall(w http.ResponseWriter, spiffeAuthZ string, backendAddress string) {
	w.Header().Set("Content-Type", "text/html")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create X509Source: %v", err), http.StatusInternalServerError)
		return
	}
	defer source.Close()

	// Allowed SPIFFE ID
	serverID := spiffeid.RequireFromString(spiffeAuthZ)

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID.
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

	// Retrieve the server SPIFFE ID from the connection.
	serverSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*resp.TLS)
	if err != nil {
		http.Error(w, fmt.Sprintf("Wasn't able to determine the SPIFFE ID of the server: %v", err), http.StatusInternalServerError)
		return
	}

	// Showcase the retrieved information and send it back to the customer.
	fmt.Fprintf(w, "<p>Got a response from: %s</p>", serverSPIFFEID.String())
	fmt.Fprintf(w, "<p>Server says: %q</p>", body)
}
