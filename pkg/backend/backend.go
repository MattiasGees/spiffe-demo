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

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type BackendService struct {
	spiffeAuthz   string
	serverAddress string
}

func StartServer(spiffeAuthz, serverAddress string) {
	backendService := BackendService{
		spiffeAuthz:   spiffeAuthz,
		serverAddress: serverAddress,
	}

	if err := backendService.run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func (b *BackendService) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up a `/` resource handler
	http.HandleFunc("/", b.rootHandler)

	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket.
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return fmt.Errorf("unable to create X509Source: %w", err)
	}
	defer source.Close()

	// Allowed SPIFFE ID
	clientID := spiffeid.RequireFromString(b.spiffeAuthz)

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID `spiffe://example.org/client`
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))
	server := &http.Server{
		Addr:              b.serverAddress,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: time.Second * 10,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (b *BackendService) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received from %s", r.RemoteAddr)
	currentTime := time.Now()
	formattedTime := currentTime.Format("02/01/06 15:04:05")
	text := fmt.Sprintf("%s: Successfully connected to the backend service!!!", formattedTime)
	_, _ = io.WriteString(w, text)
	requestorSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*r.TLS)
	if err != nil {
		log.Printf("Wasn't able to determine the SPIFFE ID of the requestor: %v", err)
	}
	log.Printf("Responded to %s with the following message: %s", requestorSPIFFEID.String(), text)
}
