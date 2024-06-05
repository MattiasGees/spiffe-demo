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

	"github.com/spf13/cobra"
)

var (
	spiffeAuthz   string
	serverAddress string
)

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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVarP(&spiffeAuthz, "authorized-spiffe", "a", "", "The SPIFFE Identity that is authorized to talk to/from this service")
	rootCmd.PersistentFlags().StringVarP(&serverAddress, "server-address", "l", "127.0.0.1:8080", "How do we want to expose our server")
}
