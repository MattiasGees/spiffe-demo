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
}

func StartServer(spiffeAuthz, serverAddress, backendService, s3Bucket, s3Filepath, awsRegion, spiffeAuthzHTTPBackend, HTTPBackendService, postgreSQLHost string) {
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
	}

	if err := customerService.run(); err != nil {
		log.Fatal(err)
	}
}

func (c *CustomerService) run() error {
	http.HandleFunc("/", c.webpageHandler)
	http.HandleFunc("/mtls", c.mtlsHandler)
	http.HandleFunc("/spifferetriever", c.spiffeRetriever)
	http.HandleFunc("/aws", c.awsRetrievalHandler)
	http.HandleFunc("/aws/put", c.awsPutHandler)
	http.HandleFunc("/httpbackend", c.httpBackendHandler)
	http.HandleFunc("/postgresql", c.postgreSQLRetrievalHandler)
	http.HandleFunc("/postgresql/put", c.postgreSQLPutHandler)

	log.Printf("Starting server at %s", c.serverAddress)

	if err := http.ListenAndServe(c.serverAddress, nil); err != nil {
		return err
	}

	return nil
}
