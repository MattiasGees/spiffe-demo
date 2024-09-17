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
	gcpBucketName          = os.Getenv("BUCKET_NAME")
	gcpProjectName         = os.Getenv("PROJECT_NAME")
	gcpJWTAudience         = os.Getenv("JWT_AUDIENCE")
	gcpSTSAudience         = os.Getenv("STS_AUDIENCE")
	gcpTokenURL            = os.Getenv("TOKEN_URL")
	gcpServiceAccountEmail = os.Getenv("SERVICE_ACCOUNT_EMAIL")
	gcpImpersonateScope    = os.Getenv("IMPERSONATE_SCOPE")
)

// GCPPutHandler writes data to a GCS bucket.
func GCPPutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client, err := createGCPStorageClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	obj := client.Bucket(gcpBucketName).Object("Hello")
	wc := obj.NewWriter(ctx)
	if _, err := wc.Write([]byte("world")); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to bucket: %v", err), http.StatusInternalServerError)
		return
	}
	if err := wc.Close(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to close writer: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Successfully wrote 'world' to 'Hello' in the GCP bucket.")
}

// GCPReadHandler reads data from a GCS bucket.
func GCPReadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client, err := createGCPStorageClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	obj := client.Bucket(gcpBucketName).Object("Hello")
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

func createGCPStorageClient(ctx context.Context) (*storage.Client, error) {
	// Retrieve JWT from SPIRE
	jwtToken, err := getGCPJWTToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT token: %v", err)
	}

	// Create a token source that exchanges the SPIRE JWT for a GCP access token
	tokenSource := oauth2.ReuseTokenSource(nil, &gcpSTSTokenSource{
		ctx:                 ctx,
		jwtToken:            jwtToken,
		tokenURL:            gcpTokenURL,
		stsAudience:         gcpSTSAudience,
		stsScope:            "https://www.googleapis.com/auth/cloud-platform",
		impersonateScope:    gcpImpersonateScope,
		serviceAccountEmail: gcpServiceAccountEmail,
	})

	client, err := storage.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return client, nil
}

func getGCPJWTToken(ctx context.Context) (string, error) {
	workloadClient, err := workloadapi.New(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create workload API client: %v", err)
	}
	defer workloadClient.Close()

	params := jwtsvid.Params{
		Audience: gcpJWTAudience,
	}

	jwtSVID, err := workloadClient.FetchJWTSVID(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWT SVID: %v", err)
	}

	jwtToken := jwtSVID.Marshal()
	log.Printf("Fetched GCP JWT Token: %s", jwtToken)

	return jwtToken, nil
}

type gcpSTSTokenSource struct {
	ctx                 context.Context
	jwtToken            string
	tokenURL            string
	stsAudience         string
	stsScope            string
	impersonateScope    string
	serviceAccountEmail string
}

func (s *gcpSTSTokenSource) Token() (*oauth2.Token, error) {
	stsAccessToken, err := s.exchangeToken()
	if err != nil {
		return nil, err
	}

	impToken, err := s.impersonateServiceAccount(stsAccessToken)
	if err != nil {
		return nil, err
	}

	return impToken, nil
}

func (s *gcpSTSTokenSource) exchangeToken() (string, error) {
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
		Scope:              s.stsScope,
		Audience:           s.stsAudience,
		SubjectToken:       s.jwtToken,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal STS request: %v", err)
	}

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

func (s *gcpSTSTokenSource) impersonateServiceAccount(stsAccessToken string) (*oauth2.Token, error) {
	impersonationURL := fmt.Sprintf("https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken", s.serviceAccountEmail)

	impReqBody := struct {
		Scope []string `json:"scope"`
	}{
		Scope: []string{s.impersonateScope},
	}

	impBody, err := json.Marshal(impReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal impersonation request: %v", err)
	}

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
