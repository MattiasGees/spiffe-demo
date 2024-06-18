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
	"log"
	"net/http"
)

func (c *CustomerService) httpBackendHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling a request in the http Backend Handler from %s", r.RemoteAddr)
	mTLSCall(w, c.spiffeAuthzHTTPBackend, c.HTTPBackendService)

}
