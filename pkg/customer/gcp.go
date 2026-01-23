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

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// gcpPutHandler writes data to a GCS bucket.
func (c *CustomerService) gcpPutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bucketName := c.getGCPBucketName()
	filePath := c.getGCPFilePath()

	client, err := c.createGCPStorageClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object(filePath)
	wc := obj.NewWriter(ctx)
	if _, err := wc.Write([]byte("world")); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to bucket: %v", err), http.StatusInternalServerError)
		return
	}
	if err := wc.Close(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to close writer: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully wrote 'world' to '%s' in the GCP bucket.", filePath)
}

// gcpReadHandler reads data from a GCS bucket.
func (c *CustomerService) gcpReadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bucketName := c.getGCPBucketName()
	filePath := c.getGCPFilePath()

	client, err := c.createGCPStorageClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage client: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object(filePath)
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

	fmt.Fprintf(w, "Content of '%s': %s", filePath, data)
}

// getGCPBucketName returns the GCP bucket name.
func (c *CustomerService) getGCPBucketName() string {
	return c.gcpBucket
}

// getGCPFilePath returns the GCP file path.
func (c *CustomerService) getGCPFilePath() string {
	return c.gcpFilePath
}

// getGCPProxyURL returns the GCP proxy URL.
func (c *CustomerService) getGCPProxyURL() string {
	return c.gcpProxyURL
}

func (c *CustomerService) createGCPStorageClient(ctx context.Context) (*storage.Client, error) {
	// Call the spiffe-gcp-proxy to get the access token
	accessToken, err := c.getGCPAccessTokenFromProxy(ctx)
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

func (c *CustomerService) getGCPAccessTokenFromProxy(ctx context.Context) (string, error) {
	proxyURL := c.getGCPProxyURL()

	// Construct the full URL for the token request
	tokenURL := fmt.Sprintf("%s/computeMetadata/v1/instance/service-accounts/default/token", proxyURL)

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
