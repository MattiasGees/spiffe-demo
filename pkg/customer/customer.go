package customer

import (
	"log"
	"net/http"
)

type CustomerService struct {
	spiffeAuthz    string
	serverAddress  string
	backendService string
	s3Bucket       string
	s3Filepath     string
	awsRegion      string
}

func StartServer(spiffeAuthz, serverAddress, backendService, s3Bucket, s3Filepath, awsRegion string) {
	customerService := CustomerService{
		spiffeAuthz:    spiffeAuthz,
		serverAddress:  serverAddress,
		backendService: backendService,
		s3Bucket:       s3Bucket,
		s3Filepath:     s3Filepath,
		awsRegion:      awsRegion,
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

	log.Printf("Starting server at %s", c.serverAddress)

	if err := http.ListenAndServe(c.serverAddress, nil); err != nil {
		return err
	}

	return nil
}
