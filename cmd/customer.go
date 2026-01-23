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
	"github.com/mattiasgees/spiffe-demo/pkg/config"
	"github.com/mattiasgees/spiffe-demo/pkg/customer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// customerCmd represents the customer command
var customerCmd = &cobra.Command{
	Use:   "customer",
	Short: "A simple customer service",
	Long: `The customer service is the endpoints that serves requests to customers.
	It connects to the backend service and relays the message back to the customer`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.CustomerConfig{
			Backend: config.BackendConfig{
				ServiceURL: viper.GetString("customer.backend.service_url"),
			},
			HTTPBackend: config.HTTPBackendConfig{
				ServiceURL: viper.GetString("customer.http_backend.service_url"),
				SpiffeID:   viper.GetString("customer.http_backend.spiffe_id"),
			},
			AWS: config.AWSConfig{
				Bucket:   viper.GetString("customer.aws.bucket"),
				FilePath: viper.GetString("customer.aws.file_path"),
			},
			GCP: config.GCPConfig{
				Bucket:   viper.GetString("customer.gcp.bucket"),
				FilePath: viper.GetString("customer.gcp.file_path"),
				ProxyURL: viper.GetString("customer.gcp.proxy_url"),
			},
			PostgreSQL: config.PostgreSQLConfig{
				Host: viper.GetString("customer.postgresql.host"),
				User: viper.GetString("customer.postgresql.user"),
			},
		}

		customer.StartServer(
			viper.GetString("spiffe.authorized_id"),
			viper.GetString("server.address"),
			cfg,
		)
	},
}

func init() {
	rootCmd.AddCommand(customerCmd)

	// Backend service flags
	customerCmd.Flags().StringP("backend-service", "b", "https://localhost:8080", "Location on where to reach the backend service")
	viper.BindPFlag("customer.backend.service_url", customerCmd.Flags().Lookup("backend-service"))
	viper.SetDefault("customer.backend.service_url", "https://localhost:8080")

	// HTTP backend service flags
	customerCmd.Flags().String("httpbackend-service", "https://localhost:8080", "Location on where to reach the HTTP backend service")
	viper.BindPFlag("customer.http_backend.service_url", customerCmd.Flags().Lookup("httpbackend-service"))
	viper.SetDefault("customer.http_backend.service_url", "https://localhost:8080")

	customerCmd.Flags().String("authorized-spiffe-httpbackend", "", "SPIFFE Identity of the HTTP backend service")
	viper.BindPFlag("customer.http_backend.spiffe_id", customerCmd.Flags().Lookup("authorized-spiffe-httpbackend"))

	// AWS flags
	customerCmd.Flags().String("aws-bucket", "", "AWS S3 bucket name")
	viper.BindPFlag("customer.aws.bucket", customerCmd.Flags().Lookup("aws-bucket"))

	customerCmd.Flags().String("aws-file-path", "testfile", "Path to the file in the AWS S3 bucket")
	viper.BindPFlag("customer.aws.file_path", customerCmd.Flags().Lookup("aws-file-path"))
	viper.SetDefault("customer.aws.file_path", "testfile")

	// GCP flags
	customerCmd.Flags().String("gcp-bucket", "", "GCP GCS bucket name")
	viper.BindPFlag("customer.gcp.bucket", customerCmd.Flags().Lookup("gcp-bucket"))

	customerCmd.Flags().String("gcp-file-path", "Hello", "Path to the file in the GCP GCS bucket")
	viper.BindPFlag("customer.gcp.file_path", customerCmd.Flags().Lookup("gcp-file-path"))
	viper.SetDefault("customer.gcp.file_path", "Hello")

	customerCmd.Flags().String("gcp-proxy-url", "http://localhost:8081", "URL of the spiffe-gcp-proxy sidecar")
	viper.BindPFlag("customer.gcp.proxy_url", customerCmd.Flags().Lookup("gcp-proxy-url"))
	viper.SetDefault("customer.gcp.proxy_url", "http://localhost:8081")

	// PostgreSQL flags
	customerCmd.Flags().String("postgresql-host", "", "Hostname of postgreSQL")
	viper.BindPFlag("customer.postgresql.host", customerCmd.Flags().Lookup("postgresql-host"))

	customerCmd.Flags().String("postgresql-user", "", "User to connect to postgreSQL")
	viper.BindPFlag("customer.postgresql.user", customerCmd.Flags().Lookup("postgresql-user"))
}
