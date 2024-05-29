package backend

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type BackendService struct {
	socketPath    string
	spiffeAuthz   string
	serverAddress string
}

func StartServer(socketPath, spiffeAuthz, serverAddress string) {
	backendService := BackendService{
		socketPath:    socketPath,
		spiffeAuthz:   spiffeAuthz,
		serverAddress: serverAddress,
	}

	if err := backendService.run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
func (b *BackendService) run(ctx context.Context) error {
	// Allowed SPIFFE ID
	clientID := spiffeid.RequireFromString(b.spiffeAuthz)

	// Creates a TLS listener
	// The server expects the client to present an SVID with the spiffeID: 'spiffe://example.org/client'
	//
	// An alternative when creating Listen is using `spiffetls.Listen` that uses environment variable `SPIFFE_ENDPOINT_SOCKET`
	listener, err := spiffetls.ListenWithMode(ctx, "tcp", b.serverAddress,
		spiffetls.MTLSServerWithSourceOptions(
			tlsconfig.AuthorizeID(clientID),
			workloadapi.WithClientOptions(workloadapi.WithAddr(b.socketPath)),
		))
	if err != nil {
		return fmt.Errorf("unable to create TLS listener: %w", err)
	}
	defer listener.Close()

	// Handle connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}
		go b.handleConnection(conn)
	}
}

func (b *BackendService) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read incoming data into buffer
	req, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("Error reading incoming data: %v", err)
		return
	}
	log.Printf("Client says: %q", req)

	// Send a response back to the client
	if _, err = conn.Write([]byte("Hello client\n")); err != nil {
		log.Printf("Unable to send response: %v", err)
		return
	}
}
