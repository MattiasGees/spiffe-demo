package customer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
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

func (c *CustomerService) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the rootHandler from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "text/html")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create X509Source: %v", err), http.StatusInternalServerError)
	}
	defer source.Close()

	// Allowed SPIFFE ID
	serverID := spiffeid.RequireFromString(c.spiffeAuthz)

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID `spiffe://example.org/server`
	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeID(serverID))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Get(c.backendService)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to %q: %v", c.backendService, err), http.StatusInternalServerError)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read body: %v", err), http.StatusInternalServerError)
	}

	serverSPIFFEID, err := spiffetls.PeerIDFromConnectionState(*resp.TLS)
	if err != nil {
		http.Error(w, fmt.Sprintf("Wasn't able to determine the SPIFFE ID of the server: %v", err), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "<p>Got a response from: %s</p>", serverSPIFFEID.String())
	fmt.Fprintf(w, "<p>Server says: %q</p>", body)
}

func (c *CustomerService) spiffeRetriever(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the SPIFFE Retriever from %s", r.RemoteAddr)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	client, err := workloadapi.New(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create workload API client: %v", err), http.StatusInternalServerError)
	}
	defer client.Close()

	x509SVIDs, err := client.FetchX509SVIDs(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch X.509 SVIDs: %v", err), http.StatusInternalServerError)
	}

	var certificates []CertificateDetails
	for _, x509SVID := range x509SVIDs {

		for _, cert := range x509SVID.Certificates {
			details := CertificateDetails{
				Issuer:             cert.Issuer.String(),
				Subject:            cert.Subject.String(),
				NotBefore:          cert.NotBefore.String(),
				NotAfter:           cert.NotAfter.String(),
				SerialNumber:       cert.SerialNumber.String(),
				SignatureAlgorithm: cert.SignatureAlgorithm.String(),
				PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
				Version:            cert.Version,
				URIs:               extractURIs(cert),
				Extensions:         extractExtensions(cert.Extensions),
			}

			certificates = append(certificates, details)
		}
	}

	JWTBundles, err := client.FetchJWTBundles(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch JWT Bundles: %v", err), http.StatusInternalServerError)
	}

	var bundles []JWTBundle
	for _, jwtbundle := range JWTBundles.Bundles() {
		var bundle JWTBundle
		jwt, err := jwtbundle.Marshal()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to marshal JWT Bundle: %v", err), http.StatusInternalServerError)
		}
		err = json.Unmarshal(jwt, &bundle)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusInternalServerError)
		}
		bundles = append(bundles, bundle)
	}

	tmpl, err := template.New("cert").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating template: %v", err), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		Certificates: certificates,
		Bundles:      bundles,
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

func (c *CustomerService) awsRetrievalHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the AWS Retrieval Handler from %s", r.RemoteAddr)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(c.awsRegion),
	})
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.s3Bucket),
		Key:    aws.String(c.s3Filepath),
	})
	if err != nil {
		http.Error(w, "Failed to get object", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read object content", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("display").Parse(`
		<html>
		<head><title>S3 File Content</title></head>
		<body>
				<h1>S3 File Content</h1>
				<pre>{{.}}</pre>
		</body>
		</html>
`))

	err = tmpl.Execute(w, string(content))
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (c *CustomerService) awsPutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the AWS Put Handler from %s", r.RemoteAddr)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(c.awsRegion),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session, %v", err), http.StatusInternalServerError)

	}

	svc := s3.New(sess)
	reader := bytes.NewReader([]byte("This is a test to write to an S3 bucket"))

	result, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(c.s3Bucket),
		Key:    aws.String(c.s3Filepath),
		Body:   reader,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to upload %q to %q, %v", c.s3Filepath, c.s3Bucket, err), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "Successfully uploaded %q to %q\n", c.s3Filepath, c.s3Bucket)
	fmt.Fprintf(w, "The uploaded content is: %s", result)
}
