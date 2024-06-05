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
	"github.com/mattiasgees/spiffe-demo/pkg/httpservice"
	"github.com/spf13/cobra"
)

// httpserviceCmd represents the httpservice command
var httpserviceCmd = &cobra.Command{
	Use:   "httpservice",
	Short: "A backend service that only can expose as an HTTP Service",
	Long: `The point of this demo is that we want to showcase how an HTTP service
	can be put behind an Envoy proxy and still do zero-trust wih SPIFFE`,
	Run: func(cmd *cobra.Command, args []string) {
		httpservice.StartServer(serverAddress)

	},
}

func init() {
	rootCmd.AddCommand(httpserviceCmd)
}
