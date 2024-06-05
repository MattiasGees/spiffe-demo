package customer

import (
	"log"
	"net/http"
)

func (c *CustomerService) httpBackendHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the http Backend Handler from %s", r.RemoteAddr)
	mTLSCall(w, r, c.spiffeAuthzHTTPBackend, c.HTTPBackendService)

}
