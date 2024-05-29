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

type CustomerService struct {
	socketPath     string
	spiffeAuthz    string
	serverAddress  string
	backendService string
}

func StartServer(socketPath, spiffeAuthz, serverAddress, backendService string) {
	customerService := CustomerService{
		socketPath:     socketPath,
		spiffeAuthz:    spiffeAuthz,
		serverAddress:  serverAddress,
		backendService: backendService,
	}

	if err := customerService.run(); err != nil {
		log.Fatal(err)
	}
}

func (c *CustomerService) run() error {
	http.HandleFunc("/", c.rootHandler)
	http.HandleFunc("/spifferetriever", c.spiffeRetriever)

	log.Printf("Starting server at %s", c.serverAddress)

	if err := http.ListenAndServe(c.serverAddress, nil); err != nil {
		return err
	}

	return nil
}

func (c *CustomerService) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the rootHandler from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "text/html")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(c.socketPath)))
	if err != nil {
		log.Printf("unable to create X509Source: %v", err)
	}
	defer source.Close()

	// Allowed SPIFFE ID
	serverID := spiffeid.RequireFromString(c.spiffeAuthz)

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID `spiffe://example.org/server`
	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeID(serverID))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	res, err := client.Get(c.backendService)
	if err != nil {
		log.Printf("error connecting to %q: %v", c.backendService, err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("unable to read body: %v", err)
	}

	serverSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*res.TLS)
	if err != nil {
		log.Printf("Wasn't able to determine the SPIFFE ID of the server: %v", err)
	}

	fmt.Fprintf(w, "<p>Got a response from: %s</p>", serverSPIFFEID.String())
	fmt.Fprintf(w, "<p>Server says: %q</p>", body)
}

func (c *CustomerService) spiffeRetriever(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "To be implememted")
}
