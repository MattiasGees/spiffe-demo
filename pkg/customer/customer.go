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
}

func StartServer(spiffeAuthz, serverAddress, backendService, s3Bucket, s3Filepath, awsRegion, spiffeAuthzHTTPBackend, HTTPBackendService string) {
	customerService := CustomerService{
		spiffeAuthz:            spiffeAuthz,
		serverAddress:          serverAddress,
		backendService:         backendService,
		s3Bucket:               s3Bucket,
		s3Filepath:             s3Filepath,
		awsRegion:              awsRegion,
		spiffeAuthzHTTPBackend: spiffeAuthzHTTPBackend,
		HTTPBackendService:     HTTPBackendService,
	}

	if err := customerService.run(); err != nil {
		log.Fatal(err)
	}
}

func (c *CustomerService) run() error {
	http.HandleFunc("/", c.rootHandler)
	http.HandleFunc("/spifferetriever", c.spiffeRetriever)
	http.HandleFunc("/aws", c.awsRetrievalHandler)
	http.HandleFunc("/aws/put", c.awsPutHandler)
	http.HandleFunc("/httpbackend", c.httpBackendHandler)

	log.Printf("Starting server at %s", c.serverAddress)

	if err := http.ListenAndServe(c.serverAddress, nil); err != nil {
		return err
	}

	return nil
}
