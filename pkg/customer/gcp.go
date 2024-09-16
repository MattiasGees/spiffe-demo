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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

var (
	bucketName          = os.Getenv("BUCKET_NAME")
	projectName         = os.Getenv("PROJECT_NAME")
	jwtAudience         = os.Getenv("JWT_AUDIENCE")
	stsAudience         = os.Getenv("STS_AUDIENCE")
	tokenURL            = os.Getenv("TOKEN_URL")
	serviceAccountEmail = os.Getenv("SERVICE_ACCOUNT_EMAIL")
	impersonateScope    = os.Getenv("IMPERSONATE_SCOPE")
)

func main() {
	if bucketName == "" || projectName == "" || jwtAudience == "" || stsAudience == "" || serviceAccountEmail == "" {
		log.Fatal("Environment variables BUCKET_NAME, PROJECT_NAME, JWT_AUDIENCE, STS_AUDIENCE, and SERVICE_ACCOUNT_EMAIL must be set.")
	}

	// Set default values if not set
	if tokenURL == "" {
		tokenURL = "https://sts.googleapis.com/v1/token"
	}
	if impersonateScope == "" {
		impersonateScope = "https://www.googleapis.com/auth/devstorage.read_write"
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/write", handleWrite)
	http.HandleFunc("/read", handleRead)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>GCS SPIFFE Application</title>
    </head>
    <body>
        <h1>GCS SPIFFE Application</h1>
        <form action="/write" method="post">
            <button type="submit">Write to Bucket</button>
        </form>
        <form action="/read" method="post">
            <button type="submit">Read from Bucket</button>
        </form>
    </body>
    </html>
    `
	fmt.Fprint(w, tmpl)
}

func handleWrite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client, err := createStorageClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object("Hello")
	wc := obj.NewWriter(ctx)
	if _, err := wc.Write([]byte("world")); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to bucket: %v", err), http.StatusInternalServerError)
		return
	}
	if err := wc.Close(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to close writer: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Successfully wrote 'world' to 'Hello' in the bucket.")
}

func handleRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client, err := createStorageClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object("Hello")
	rc, err := obj.NewReader(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read from bucket: %v", err), http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read data: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Content of 'Hello': %s", data)
}

func createStorageClient(ctx context.Context) (*storage.Client, error) {
	// Retrieve JWT from SPIRE
	jwtToken, err := getJWTToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT token: %v", err)
	}

	// Create a token source that exchanges the SPIRE JWT for a GCP access token
	tokenSource := oauth2.ReuseTokenSource(nil, &stsTokenSource{
		ctx:                 ctx,
		jwtToken:            jwtToken,
		tokenURL:            tokenURL,
		stsAudience:         stsAudience,
		stsScope:            "https://www.googleapis.com/auth/cloud-platform", // Scope for STS token
		impersonateScope:    impersonateScope,                                 // Scope for final access token
		serviceAccountEmail: serviceAccountEmail,
	})

	// Create the storage client with the token source
	client, err := storage.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return client, nil
}

func getJWTToken(ctx context.Context) (string, error) {
	// Create a Workload API client
	workloadClient, err := workloadapi.New(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create workload API client: %v", err)
	}
	defer workloadClient.Close()

	// Set the audience to the resource you're accessing
	params := jwtsvid.Params{
		Audience: jwtAudience,
	}

	// Fetch the JWT SVID
	jwtSVID, err := workloadClient.FetchJWTSVID(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWT SVID: %v", err)
	}

	jwtToken := jwtSVID.Marshal()

	// Log the JWT token (only for debugging; remove in production)
	log.Printf("Fetched JWT Token: %s", jwtToken)

	return jwtToken, nil
}

// stsTokenSource exchanges the SPIRE JWT for a GCP access token using the STS endpoint
type stsTokenSource struct {
	ctx                 context.Context
	jwtToken            string
	tokenURL            string
	stsAudience         string
	stsScope            string
	impersonateScope    string
	serviceAccountEmail string
}

func (s *stsTokenSource) Token() (*oauth2.Token, error) {
	// Step 1: Exchange JWT for STS access token
	stsAccessToken, err := s.exchangeToken()
	if err != nil {
		return nil, err
	}

	// Step 2: Impersonate the service account using the STS access token
	impToken, err := s.impersonateServiceAccount(stsAccessToken)
	if err != nil {
		return nil, err
	}

	return impToken, nil
}

func (s *stsTokenSource) exchangeToken() (string, error) {
	// Build the STS request
	reqBody := struct {
		GrantType          string `json:"grant_type"`
		RequestedTokenType string `json:"requested_token_type"`
		Scope              string `json:"scope"`
		Audience           string `json:"audience"`
		SubjectToken       string `json:"subject_token"`
		SubjectTokenType   string `json:"subject_token_type"`
	}{
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		Scope:              s.stsScope, // Use stsScope here
		Audience:           s.stsAudience,
		SubjectToken:       s.jwtToken,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
	}

	// Marshal the request body
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal STS request: %v", err)
	}

	// Make the HTTP request to the STS endpoint
	client := &http.Client{}
	req, err := http.NewRequest("POST", s.tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create STS request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req.WithContext(s.ctx))
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, responseBody)
	}

	var stsTokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stsTokenResp); err != nil {
		return "", fmt.Errorf("failed to decode STS token response: %v", err)
	}

	return stsTokenResp.AccessToken, nil
}

func (s *stsTokenSource) impersonateServiceAccount(stsAccessToken string) (*oauth2.Token, error) {
	// Build the impersonation request
	impersonationURL := fmt.Sprintf("https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken", s.serviceAccountEmail)

	impReqBody := struct {
		Scope []string `json:"scope"`
	}{
		Scope: []string{s.impersonateScope}, // Use impersonateScope here
	}

	impBody, err := json.Marshal(impReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal impersonation request: %v", err)
	}

	// Make the HTTP request to the impersonation endpoint
	client := &http.Client{}
	impReq, err := http.NewRequest("POST", impersonationURL, bytes.NewReader(impBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create impersonation request: %v", err)
	}
	impReq.Header.Set("Content-Type", "application/json")
	impReq.Header.Set("Authorization", "Bearer "+stsAccessToken)

	impResp, err := client.Do(impReq.WithContext(s.ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to impersonate service account: %v", err)
	}
	defer impResp.Body.Close()

	if impResp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(impResp.Body)
		return nil, fmt.Errorf("service account impersonation failed with status %d: %s", impResp.StatusCode, responseBody)
	}

	var impTokenResp struct {
		AccessToken string `json:"accessToken"`
		ExpireTime  string `json:"expireTime"`
	}

	if err := json.NewDecoder(impResp.Body).Decode(&impTokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode impersonation token response: %v", err)
	}

	expiry, err := time.Parse(time.RFC3339, impTokenResp.ExpireTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token expiry time: %v", err)
	}

	token := &oauth2.Token{
		AccessToken: impTokenResp.AccessToken,
		TokenType:   "Bearer",
		Expiry:      expiry,
	}

	return token, nil
}
