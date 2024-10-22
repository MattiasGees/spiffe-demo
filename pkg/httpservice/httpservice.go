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
package httpservice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type HTTPService struct {
	serverAddress string
}

// Main function that creates the httpbackend server and starts it. This is called from the CLI.
func StartServer(serverAddress string) {
	HTTPService := HTTPService{
		serverAddress: serverAddress,
	}

	if err := HTTPService.run(); err != nil {
		log.Fatal(err)
	}
}

// This gets called from the main function and actually starts an HTTP server.
func (h *HTTPService) run() error {
	// Set up a `/` resource handler
	http.HandleFunc("/", h.rootHandler)

	log.Printf("Starting server at %s", h.serverAddress)

	// Serve the HTTP server
	if err := http.ListenAndServe(h.serverAddress, nil); err != nil {
		return err
	}

	return nil
}

// function that handles calls to `/`. This will just respond with a simple message and the date and time.
func (h *HTTPService) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received from %s", r.RemoteAddr)
	currentTime := time.Now()
	formattedTime := currentTime.Format("02/01/06 15:04:05")
	text := fmt.Sprintf("%s: Successfully connected to the HTTP service!!!", formattedTime)
	_, _ = io.WriteString(w, text)
}
