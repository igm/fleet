package api

import (
	"net/http"

	"github.com/coreos/fleet/client"
	"github.com/coreos/fleet/log"
	"github.com/coreos/fleet/registry"
)

func NewServeMux(reg registry.Registry) http.Handler {
	sm := http.NewServeMux()
	cAPI := &client.RegistryClient{reg}

	prefix := "/v1-alpha"
	wireUpDiscoveryResource(sm, prefix)
	wireUpMachinesResource(sm, prefix, cAPI)
	wireUpStateResource(sm, prefix, cAPI)
	wireUpUnitsResource(sm, prefix, cAPI)

	sm.HandleFunc(prefix, methodNotAllowedHandler)
	sm.HandleFunc("/", baseHandler)

	lm := &loggingMiddleware{sm}

	return lm
}

type loggingMiddleware struct {
	next http.Handler
}

func (lm *loggingMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.V(1).Infof("HTTP %s %v", req.Method, req.URL)
	lm.next.ServeHTTP(rw, req)
}

func methodNotAllowedHandler(rw http.ResponseWriter, req *http.Request) {
	sendError(rw, http.StatusMethodNotAllowed, nil)
}

func baseHandler(rw http.ResponseWriter, req *http.Request) {
	var code int
	if req.URL.Path == "/" {
		code = http.StatusMethodNotAllowed
	} else {
		code = http.StatusNotFound
	}

	sendError(rw, code, nil)
}
