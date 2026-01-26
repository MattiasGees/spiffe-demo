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
)

var ledgerCmd = &cobra.Command{
	Use:   "ledger",
	Short: "TrustBank ledger service",
	Long: `The ledger service provides a secure banking API for the TrustBank demo.

It exposes REST endpoints for managing accounts and transfers, secured with
SPIFFE mTLS authentication. Only clients with authorized SPIFFE IDs can connect.

Endpoints:
  GET  /api/accounts         - List all accounts with balances
  GET  /api/accounts/{id}    - Get a specific account
  POST /api/transfers        - Create a transfer between accounts
  GET  /api/transactions     - List all transactions
  GET  /api/transactions/{id} - Get a specific transaction
  GET  /health               - Health check endpoint`,
	Run: func(cmd *cobra.Command, args []string) {
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
			// Use PostgreSQL store
			cfg := ledger.PostgresConfig{
				Host:        viper.GetString("ledger.postgresql.host"),
				Port:        viper.GetInt("ledger.postgresql.port"),
				User:        viper.GetString("ledger.postgresql.user"),
				Database:    viper.GetString("ledger.postgresql.database"),
				SSLMode:     viper.GetString("ledger.postgresql.ssl_mode"),
				SSLCert:     viper.GetString("ledger.postgresql.ssl_cert"),
				SSLKey:      viper.GetString("ledger.postgresql.ssl_key"),
				SSLRootCert: viper.GetString("ledger.postgresql.ssl_root_cert"),
			}

			pgStore, err := ledger.NewPostgresStore(context.Background(), cfg)
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
	ledgerCmd.Flags().Bool("use-mock", false, "Use in-memory mock store instead of PostgreSQL")
	ledgerCmd.Flags().String("postgresql-host", "localhost", "PostgreSQL host")
	ledgerCmd.Flags().Int("postgresql-port", 5432, "PostgreSQL port")
	ledgerCmd.Flags().String("postgresql-user", "spiffe-demo-ledger", "PostgreSQL user")
	ledgerCmd.Flags().String("postgresql-database", "trustbank", "PostgreSQL database")
	ledgerCmd.Flags().String("postgresql-ssl-mode", "require", "PostgreSQL SSL mode")
	ledgerCmd.Flags().String("postgresql-ssl-cert", "", "Path to client SSL certificate")
	ledgerCmd.Flags().String("postgresql-ssl-key", "", "Path to client SSL key")
	ledgerCmd.Flags().String("postgresql-ssl-root-cert", "", "Path to SSL root certificate")

	// Bind flags to viper
	viper.BindPFlag("ledger.use_mock", ledgerCmd.Flags().Lookup("use-mock"))
	viper.BindPFlag("ledger.postgresql.host", ledgerCmd.Flags().Lookup("postgresql-host"))
	viper.BindPFlag("ledger.postgresql.port", ledgerCmd.Flags().Lookup("postgresql-port"))
	viper.BindPFlag("ledger.postgresql.user", ledgerCmd.Flags().Lookup("postgresql-user"))
	viper.BindPFlag("ledger.postgresql.database", ledgerCmd.Flags().Lookup("postgresql-database"))
	viper.BindPFlag("ledger.postgresql.ssl_mode", ledgerCmd.Flags().Lookup("postgresql-ssl-mode"))
	viper.BindPFlag("ledger.postgresql.ssl_cert", ledgerCmd.Flags().Lookup("postgresql-ssl-cert"))
	viper.BindPFlag("ledger.postgresql.ssl_key", ledgerCmd.Flags().Lookup("postgresql-ssl-key"))
	viper.BindPFlag("ledger.postgresql.ssl_root_cert", ledgerCmd.Flags().Lookup("postgresql-ssl-root-cert"))
}
