package watcher

import (
	"log"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// x509Watcher is a sample implementation of the workloadapi.X509ContextWatcher interface
type x509Watcher struct{}

func NewX509Watcher() *x509Watcher {
	return &x509Watcher{}
}

// UpdateX509SVIDs is run every time an SVID is updated
func (x509Watcher) OnX509ContextUpdate(c *workloadapi.X509Context) {
	for _, svid := range c.SVIDs {
		pem, _, err := svid.Marshal()
		if err != nil {
			log.Fatalf("Unable to marshal X.509 SVID: %v", err)
		}

		log.Printf("SVID updated for %q: \n%s\n", svid.ID, string(pem))
	}
}

// OnX509ContextWatchError is run when the client runs into an error
func (x509Watcher) OnX509ContextWatchError(err error) {
	if status.Code(err) != codes.Canceled {
		log.Printf("OnX509ContextWatchError error: %v", err)
	}
}
