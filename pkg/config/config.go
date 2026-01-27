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
package config

// Config represents the root configuration structure.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Spiffe   SpiffeConfig   `mapstructure:"spiffe"`
	Customer CustomerConfig `mapstructure:"customer"`
	Ledger   LedgerConfig   `mapstructure:"ledger"`
}

// ServerConfig contains server-related configuration.
type ServerConfig struct {
	Address string `mapstructure:"address"`
}

// SpiffeConfig contains SPIFFE-related configuration.
type SpiffeConfig struct {
	AuthorizedID string `mapstructure:"authorized_id"`
}

// CustomerConfig contains customer service specific configuration.
type CustomerConfig struct {
	Backend     BackendConfig     `mapstructure:"backend"`
	HTTPBackend HTTPBackendConfig `mapstructure:"http_backend"`
	AWS         AWSConfig         `mapstructure:"aws"`
	GCP         GCPConfig         `mapstructure:"gcp"`
	PostgreSQL  PostgreSQLConfig  `mapstructure:"postgresql"`
}

// BackendConfig contains backend service connection configuration.
type BackendConfig struct {
	ServiceURL string `mapstructure:"service_url"`
}

// HTTPBackendConfig contains HTTP backend service configuration.
type HTTPBackendConfig struct {
	ServiceURL string `mapstructure:"service_url"`
	SpiffeID   string `mapstructure:"spiffe_id"`
}

// AWSConfig contains AWS-related configuration.
// Note: AWS region is intentionally omitted - use AWS_REGION environment variable.
type AWSConfig struct {
	Bucket   string `mapstructure:"bucket"`
	FilePath string `mapstructure:"file_path"`
}

// GCPConfig contains GCP-related configuration.
type GCPConfig struct {
	Bucket   string `mapstructure:"bucket"`
	FilePath string `mapstructure:"file_path"`
	ProxyURL string `mapstructure:"proxy_url"`
}

// PostgreSQLConfig contains PostgreSQL connection configuration.
type PostgreSQLConfig struct {
	Host string `mapstructure:"host"`
	User string `mapstructure:"user"`
}

// LedgerConfig contains ledger service specific configuration.
type LedgerConfig struct {
	UseMock    bool                    `mapstructure:"use_mock"`
	PostgreSQL LedgerPostgreSQLConfig `mapstructure:"postgresql"`
}

// LedgerPostgreSQLConfig contains PostgreSQL configuration for the ledger service.
// Note: SSL/TLS and username are handled via SPIFFE X509Source.
// The username is automatically extracted from the SVID certificate's Common Name.
type LedgerPostgreSQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}
