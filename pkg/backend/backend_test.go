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
package backend

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootHandler(t *testing.T) {
	svc := BackendService{
		spiffeAuthz:   "spiffe://example.org/test",
		serverAddress: ":8443",
	}

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err, "creating request should not fail")

	// Add a mock TLS connection state (the handler checks r.TLS)
	req.TLS = &tls.ConnectionState{}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(svc.rootHandler)

	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code, "handler should return OK status")

	// Check response body contains expected message
	body := rr.Body.String()
	assert.Contains(t, body, "Successfully connected to the backend service!!!", "response should contain success message")
}

func TestRootHandlerResponseFormat(t *testing.T) {
	svc := BackendService{
		spiffeAuthz:   "spiffe://example.org/test",
		serverAddress: ":8443",
	}

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	// Add a mock TLS connection state
	req.TLS = &tls.ConnectionState{}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(svc.rootHandler)

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()

	// Response should start with a timestamp in format "DD/MM/YY HH:MM:SS"
	// The format is: "02/01/06 15:04:05: Successfully connected..."
	parts := strings.SplitN(body, ": ", 2)
	require.Len(t, parts, 2, "response should have timestamp followed by message")

	// Verify timestamp format (DD/MM/YY HH:MM:SS)
	timestamp := parts[0]
	assert.Regexp(t, `^\d{2}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}$`, timestamp, "timestamp should match expected format")
}
