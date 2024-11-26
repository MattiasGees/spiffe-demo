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
package customer

import (
	"log"
	"net/http"
)

type CustomerService struct {
	spiffeAuthz            string
	serverAddress          string
	backendService         string
	s3Bucket               string
	s3Filepath             string
	awsRegion              string
	spiffeAuthzHTTPBackend string
	HTTPBackendService     string
	postgreSQLHost         string
	postgreSQLUser         string
}

// Main function that creates the customer server and starts it. This is called from the CLI.
func StartServer(spiffeAuthz, serverAddress, backendService, s3Bucket, s3Filepath, awsRegion, spiffeAuthzHTTPBackend, HTTPBackendService, postgreSQLHost, postgreSQLUser string) {
	customerService := CustomerService{
		spiffeAuthz:            spiffeAuthz,
		serverAddress:          serverAddress,
		backendService:         backendService,
		s3Bucket:               s3Bucket,
		s3Filepath:             s3Filepath,
		awsRegion:              awsRegion,
		spiffeAuthzHTTPBackend: spiffeAuthzHTTPBackend,
		HTTPBackendService:     HTTPBackendService,
		postgreSQLHost:         postgreSQLHost,
		postgreSQLUser:         postgreSQLUser,
	}

	if err := customerService.run(); err != nil {
		log.Fatal(err)
	}
}

// This gets called from the main function and actually starts that customer HTTP server.
func (c *CustomerService) run() error {
	// Set up all of the resource handlers.
	http.HandleFunc("/", c.webpageHandler)
	http.HandleFunc("/mtls", c.mtlsHandler)
	http.HandleFunc("/spifferetriever", c.spiffeRetriever)
	http.HandleFunc("/aws", c.awsRetrievalHandler)
	http.HandleFunc("/aws/put", c.awsPutHandler)
	http.HandleFunc("/gcp/put", GCPPutHandler)
	http.HandleFunc("/gcp", GCPReadHandler)
	http.HandleFunc("/httpbackend", c.httpBackendHandler)
	http.HandleFunc("/postgresql", c.postgreSQLRetrievalHandler)
	http.HandleFunc("/postgresql/put", c.postgreSQLPutHandler)

	log.Printf("Starting server at %s", c.serverAddress)

	// Serve the HTTP server.
	if err := http.ListenAndServe(c.serverAddress, nil); err != nil {
		return err
	}

	return nil
}
