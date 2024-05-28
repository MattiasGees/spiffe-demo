package watcher

import (
	"log"

	"github.com/spiffe/go-spiffe/v2/bundle/jwtbundle"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// jwtWatcher is a sample implementation of the workloadapi.JWTBundleWatcher interface
type jwtWatcher struct{}

func NewJWTWatcher() *jwtWatcher {
	return &jwtWatcher{}
}

// UpdateX509SVIDs is run every time a JWT Bundle is updated
func (jwtWatcher) OnJWTBundlesUpdate(bundleSet *jwtbundle.Set) {
	for _, bundle := range bundleSet.Bundles() {
		jwt, err := bundle.Marshal()
		if err != nil {
			log.Fatalf("Unable to marshal JWT Bundle : %v", err)
		}
		log.Printf("jwt bundle updated %q: %s", bundle.TrustDomain(), string(jwt))
	}
}

// OnJWTBundlesWatchError is run when the client runs into an error
func (jwtWatcher) OnJWTBundlesWatchError(err error) {
	if status.Code(err) != codes.Canceled {
		log.Printf("OnJWTBundlesWatchError error: %v", err)
	}
}
