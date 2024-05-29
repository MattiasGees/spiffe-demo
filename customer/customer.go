package customer

import (
	"bufio"
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

func (c *CustomerService) rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Allowed SPIFFE ID
	serverID := spiffeid.RequireFromString(c.spiffeAuthz)

	// Create a TLS connection.
	// The client expects the server to present an SVID with the spiffeID: 'spiffe://example.org/server'
	//
	// An alternative when creating Dial is using `spiffetls.Dial` that uses environment variable `SPIFFE_ENDPOINT_SOCKET`
	conn, err := spiffetls.DialWithMode(ctx, "tcp", c.backendService,
		spiffetls.MTLSClientWithSourceOptions(
			tlsconfig.AuthorizeID(serverID),
			workloadapi.WithClientOptions(workloadapi.WithAddr(c.socketPath)),
		))
	if err != nil {
		fmt.Fprintf(w, "unable to create TLS connection: %w", err)
	}
	defer conn.Close()

	// Send a message to the server using the TLS connection
	fmt.Fprintf(conn, "Hello server\n")

	// Read server response
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintf(w, "unable to read server response: %w", err)
	}
	fmt.Fprintf(w, "Server says: %q", status)
}

func (c *CustomerService) run() error {
	http.HandleFunc("/", c.rootHandler)

	log.Printf("Starting server at %s", c.serverAddress)

	if err := http.ListenAndServe(c.serverAddress, nil); err != nil {
		return err
	}

	return nil
}
