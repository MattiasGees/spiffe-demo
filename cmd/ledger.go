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
package cmd

import (
	"context"
	"log"

	"github.com/mattiasgees/spiffe-demo/pkg/ledger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

var ledgerCmd = &cobra.Command{
	Use:   "ledger",
	Short: "TrustBank ledger service",
	Long: `The ledger service provides a secure banking API for the TrustBank demo.

It exposes REST endpoints for managing accounts and transfers, secured with
SPIFFE mTLS authentication. Only clients with authorized SPIFFE IDs can connect.

The service uses SPIFFE X.509-SVIDs for:
  - Server mTLS authentication (incoming client connections)
  - PostgreSQL client authentication (database connections)

Endpoints:
  GET  /api/accounts         - List all accounts with balances
  GET  /api/accounts/{id}    - Get a specific account
  POST /api/transfers        - Create a transfer between accounts
  GET  /api/transactions     - List all transactions
  GET  /api/transactions/{id} - Get a specific transaction
  GET  /health               - Health check endpoint`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get configuration
		serverAddress := viper.GetString("server.address")
		spiffeAuthz := viper.GetString("spiffe.authorized_id")

		// Determine storage backend
		var store ledger.Store
		useMock := viper.GetBool("ledger.use_mock")

		if useMock {
			log.Println("Using mock store (in-memory)")
			store = ledger.NewMockStore()
		} else {
			// SPIFFE CONCEPT: Unified Identity for All Connections
			// The X509Source provides the service's SPIFFE identity which is used for:
			// 1. mTLS server authentication (clients connecting to this service)
			// 2. mTLS client authentication (this service connecting to PostgreSQL)
			// This demonstrates how a single SPIFFE identity can secure multiple
			// connection types without managing separate certificates.
			source, err := workloadapi.NewX509Source(ctx)
			if err != nil {
				log.Fatalf("Failed to create X509Source: %v", err)
			}
			defer source.Close()

			svid, err := source.GetX509SVID()
			if err != nil {
				log.Fatalf("Failed to get X509-SVID: %v", err)
			}
			log.Printf("Using SPIFFE identity: %s", svid.ID.String())

			// Use PostgreSQL store with SPIFFE authentication
			// Note: Username is automatically extracted from the SVID's Common Name
			cfg := ledger.PostgresConfig{
				Host:     viper.GetString("ledger.postgresql.host"),
				Port:     viper.GetInt("ledger.postgresql.port"),
				Database: viper.GetString("ledger.postgresql.database"),
			}

			pgStore, err := ledger.NewPostgresStore(ctx, cfg, source)
			if err != nil {
				log.Fatalf("Failed to connect to PostgreSQL: %v", err)
			}
			defer pgStore.Close()
			store = pgStore
		}

		// Start the server
		ledger.StartServer(spiffeAuthz, serverAddress, store)
	},
}

func init() {
	rootCmd.AddCommand(ledgerCmd)

	// Ledger-specific flags
	// Note: PostgreSQL username is derived from the SPIFFE SVID's Common Name
	ledgerCmd.Flags().Bool("use-mock", false, "Use in-memory mock store instead of PostgreSQL")
	ledgerCmd.Flags().String("postgresql-host", "localhost", "PostgreSQL host")
	ledgerCmd.Flags().Int("postgresql-port", 5432, "PostgreSQL port")
	ledgerCmd.Flags().String("postgresql-database", "trustbank", "PostgreSQL database")

	// Bind flags to viper
	viper.BindPFlag("ledger.use_mock", ledgerCmd.Flags().Lookup("use-mock"))
	viper.BindPFlag("ledger.postgresql.host", ledgerCmd.Flags().Lookup("postgresql-host"))
	viper.BindPFlag("ledger.postgresql.port", ledgerCmd.Flags().Lookup("postgresql-port"))
	viper.BindPFlag("ledger.postgresql.database", ledgerCmd.Flags().Lookup("postgresql-database"))
}
