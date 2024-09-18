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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

var (
	gcpBucketName = os.Getenv("BUCKET_NAME")
)

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

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
	// Call the spiffe-gcp-proxy to get the access token
	accessToken, err := getGCPAccessTokenFromProxy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token from proxy: %v", err)
	}

	// Use the access token to create the storage client
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})

	client, err := storage.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return client, nil
}

func getGCPAccessTokenFromProxy(ctx context.Context) (string, error) {
	// Get the proxy URL from an environment variable (default to localhost:8080)
	GCPproxyURL := os.Getenv("SPIFFE_GCP_PROXY_URL")
	if GCPproxyURL == "" {
		GCPproxyURL = "http://localhost:8080"
	}

	// Construct the full URL for the token request
	tokenURL := fmt.Sprintf("%s/computeMetadata/v1/instance/service-accounts/default/token", GCPproxyURL)

	// Request token from spiffe-gcp-proxy
	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get response from proxy: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("proxy returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp AccessTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode proxy response: %v", err)
	}

	return tokenResp.AccessToken, nil
}
