package customer

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type CertificateDetails struct {
	Issuer             string
	Subject            string
	NotBefore          string
	NotAfter           string
	SerialNumber       string
	SignatureAlgorithm string
	PublicKeyAlgorithm string
	Version            int
	URIs               []string
	DNSNames           []string
	Extensions         []string
}

type PageData struct {
	Certificates []CertificateDetails
	Bundles      []JWTBundle
}

type JWTKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWTBundle struct {
	Keys []JWTKey `json:"keys"`
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Certificates Details</title>
	<style>
			body { font-family: Arial, sans-serif; }
			.container { max-width: 800px; margin: auto; padding: 20px; }
			table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
			th, td { padding: 10px; border: 1px solid #ddd; }
			th { background-color: #f4f4f4; }
			.extensions { font-size: 0.9em; color: #555; }
			.cert-container { margin-bottom: 40px; }
	</style>
</head>
<body>
	<div class="container">
			<h1>Certificate Details</h1>
			{{ range $index, $cert := .Certificates }}
			<div class="cert-container">
					<h2>Certificate {{ $index }}</h2>
					<table>
							<tr><th>Issuer</th><td>{{ $cert.Issuer }}</td></tr>
							<tr><th>Subject</th><td>{{ $cert.Subject }}</td></tr>
							<tr><th>Not Before</th><td>{{ $cert.NotBefore }}</td></tr>
							<tr><th>Not After</th><td>{{ $cert.NotAfter }}</td></tr>
							<tr><th>Serial Number</th><td>{{ $cert.SerialNumber }}</td></tr>
							<tr><th>Signature Algorithm</th><td>{{ $cert.SignatureAlgorithm }}</td></tr>
							<tr><th>Public Key Algorithm</th><td>{{ $cert.PublicKeyAlgorithm }}</td></tr>
							<tr><th>Version</th><td>{{ $cert.Version }}</td></tr>
							<tr><th>URIs</th><td>{{ range $cert.URIs }}<div>{{ . }}</div>{{ end }}</td></tr>
							<tr><th>DNS Names</th><td>{{ range $cert.DNSNames }}<div>{{ . }}</div>{{ end }}</td></tr>
					</table>
					<h3>Extensions</h3>
					<ul class="extensions">
							{{ range $cert.Extensions }}
							<li>{{ . }}</li>
							{{ end }}
					</ul>
			</div>
			{{ end }}
			<h1>JWT Bundles</h1>
			{{ range $bundleIndex, $bundle := .Bundles }}
			<h2>Bundle {{ $bundleIndex }}</h2>
			{{ range $keyIndex, $key := $bundle.Keys }}
			<h3>Key {{ $keyIndex }}</h3>
			<table>
					<tr><th>Key Type</th><td>{{ $key.Kty }}</td></tr>
					<tr><th>Key ID</th><td>{{ $key.Kid }}</td></tr>
					<tr><th>Modulus</th><td><textarea readonly rows="5" style="width:100%;">{{ $key.N }}</textarea></td></tr>
					<tr><th>Exponent</th><td>{{ $key.E }}</td></tr>
			</table>
			{{ end }}
			{{ end }}
	</div>
</body>
</html>
`

func (c *CustomerService) spiffeRetriever(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the SPIFFE Retriever from %s", r.RemoteAddr)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	client, err := workloadapi.New(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create workload API client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	x509SVIDs, err := client.FetchX509SVIDs(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch X.509 SVIDs: %v", err), http.StatusInternalServerError)
		return
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
				DNSNames:           cert.DNSNames,
				Extensions:         extractExtensions(cert.Extensions),
			}

			certificates = append(certificates, details)
		}
	}

	JWTBundles, err := client.FetchJWTBundles(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch JWT Bundles: %v", err), http.StatusInternalServerError)
		return
	}

	var bundles []JWTBundle
	for _, jwtbundle := range JWTBundles.Bundles() {
		var bundle JWTBundle
		jwt, err := jwtbundle.Marshal()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to marshal JWT Bundle: %v", err), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(jwt, &bundle)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusInternalServerError)
			return
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

func extractURIs(cert *x509.Certificate) []string {
	var uris []string
	for _, uri := range cert.URIs {
		uris = append(uris, uri.String())
	}
	return uris
}

func extractExtensions(exts []pkix.Extension) []string {
	var extensionDetails []string
	for _, ext := range exts {
		extDetail := fmt.Sprintf("ID: %s, Critical: %t, Value: %x", ext.Id.String(), ext.Critical, ext.Value)
		extensionDetails = append(extensionDetails, extDetail)
	}
	return extensionDetails
}
