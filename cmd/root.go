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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spiffe-demo",
	Short: "The SPIFFE demo command line utility",
	Long: `This command line utility will be used to demo different SPIFFE capabilities.
	It can be started in different modes to resemble different applications`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml, ~/.spiffe-demo/config.yaml, /etc/spiffe-demo/config.yaml)")

	// Global flags
	rootCmd.PersistentFlags().StringP("authorized-spiffe", "a", "", "The SPIFFE Identity that is authorized to talk to/from this service")
	rootCmd.PersistentFlags().StringP("server-address", "l", "127.0.0.1:8080", "How do we want to expose our server")

	// Bind flags to viper
	viper.BindPFlag("spiffe.authorized_id", rootCmd.PersistentFlags().Lookup("authorized-spiffe"))
	viper.BindPFlag("server.address", rootCmd.PersistentFlags().Lookup("server-address"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in standard locations
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.spiffe-demo")
		viper.AddConfigPath("/etc/spiffe-demo")
	}

	// Environment variables
	viper.SetEnvPrefix("SPIFFE_DEMO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server.address", "127.0.0.1:8080")

	// Read config file if present (not an error if missing)
	viper.ReadInConfig()
}
