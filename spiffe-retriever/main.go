package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/mattiasgees/spiffe-demo/pkg/watcher"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const socketPath = "unix:///spiffe-workload-api/socket"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Wait for an os.Interrupt signal
	go waitForCtrlC(cancel)

	// Start X.509 and JWT watchers
	startWatchers(ctx)
}

func startWatchers(ctx context.Context) {
	var wg sync.WaitGroup

	// Creates a new Workload API client, connecting to provided socket path
	// Environment variable `SPIFFE_ENDPOINT_SOCKET` is used as default
	client, err := workloadapi.New(ctx, workloadapi.WithAddr(socketPath))
	if err != nil {
		log.Fatalf("Unable to create workload API client: %v", err)
	}
	defer client.Close()

	wg.Add(1)
	// Start a watcher for X.509 SVID updates
	go func() {
		defer wg.Done()
		err := client.WatchX509Context(ctx, watcher.NewX509Watcher())
		if err != nil && status.Code(err) != codes.Canceled {
			log.Fatalf("Error watching X.509 context: %v", err)
		}
	}()

	wg.Add(1)
	// Start a watcher for JWT bundle updates
	go func() {
		defer wg.Done()
		err := client.WatchJWTBundles(ctx, watcher.NewJWTWatcher())
		if err != nil && status.Code(err) != codes.Canceled {
			log.Fatalf("Error watching JWT bundles: %v", err)
		}
	}()

	wg.Wait()
}

// waitForCtrlC waits until an os.Interrupt signal is sent (ctrl + c)
func waitForCtrlC(cancel context.CancelFunc) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh

	cancel()
}
