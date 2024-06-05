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

func StartServer(serverAddress string) {
	HTTPService := HTTPService{
		serverAddress: serverAddress,
	}

	if err := HTTPService.run(); err != nil {
		log.Fatal(err)
	}
}

func (h *HTTPService) run() error {
	http.HandleFunc("/", h.rootHandler)

	log.Printf("Starting server at %s", h.serverAddress)

	if err := http.ListenAndServe(h.serverAddress, nil); err != nil {
		return err
	}

	return nil
}

func (h *HTTPService) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received from %s", r.RemoteAddr)
	currentTime := time.Now()
	formattedTime := currentTime.Format("02/01/06 15:04:05")
	text := fmt.Sprintf("%s: Successfully connected to the HTTP service!!!", formattedTime)
	_, _ = io.WriteString(w, text)
}
