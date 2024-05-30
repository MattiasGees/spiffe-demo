package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type CustomerService struct {
	socketPath     string
	spiffeAuthz    string
	serverAddress  string
	backendService string
}

func StartServer(socketPath, spiffeAuthz, serverAddress, backendService string) {
	customerService := CustomerService{
		socketPath:     socketPath,
		spiffeAuthz:    spiffeAuthz,
		serverAddress:  serverAddress,
		backendService: backendService,
	}

	if err := customerService.run(); err != nil {
		log.Fatal(err)
	}
}

func (c *CustomerService) run() error {
	http.HandleFunc("/", c.rootHandler)
	http.HandleFunc("/spifferetriever", c.spiffeRetriever)

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
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(c.socketPath)))
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

	client, err := workloadapi.New(ctx, workloadapi.WithAddr(c.socketPath))
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
