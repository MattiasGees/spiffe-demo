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
	"github.com/mattiasgees/spiffe-demo/backend"
	"github.com/spf13/cobra"
)

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "A simple backend service",
	Long: `This starts a simple backend service that will be exposes as an mTLS SPIFFE Service.
	It will validate incoming requests based on a SPIFFE identity`,
	Run: func(cmd *cobra.Command, args []string) {
		backend.StartServer(socketPath, spiffeAuthz, serverAddress)
	},
}

func init() {
	rootCmd.AddCommand(backendCmd)
}
