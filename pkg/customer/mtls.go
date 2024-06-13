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

func (c *CustomerService) mtlsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the rootHandler from %s", r.RemoteAddr)
	mTLSCall(w, c.spiffeAuthz, c.backendService)
}

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

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID `spiffe://example.org/server`
	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeID(serverID))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Get(backendAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to %q: %v", backendAddress, err), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read body: %v", err), http.StatusInternalServerError)
		return
	}

	serverSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*resp.TLS)
	if err != nil {
		http.Error(w, fmt.Sprintf("Wasn't able to determine the SPIFFE ID of the server: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<p>Got a response from: %s</p>", serverSPIFFEID.String())
	fmt.Fprintf(w, "<p>Server says: %q</p>", body)
}
