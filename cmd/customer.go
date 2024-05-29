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
	"github.com/mattiasgees/spiffe-demo/customer"
	"github.com/spf13/cobra"
)

var (
	backendService string
)

// customerCmd represents the customer command
var customerCmd = &cobra.Command{
	Use:   "customer",
	Short: "A simple customer service",
	Long: `The customer service is the endpoints that serves requests to customers.
	It connects to the backend service and relays the message back to the customer`,
	Run: func(cmd *cobra.Command, args []string) {
		customer.StartServer(socketPath, spiffeAuthz, serverAddress, backendService)
	},
}

func init() {
	rootCmd.AddCommand(customerCmd)
	customerCmd.PersistentFlags().StringVarP(&backendService, "backend-service", "b", "localhost:8080", "Location on where to reach the backend service")
}
