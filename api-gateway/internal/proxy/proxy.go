package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	targets map[string]*httputil.ReverseProxy
}

func NewProxy() *Proxy {
	return &Proxy{
		targets: make(map[string]*httputil.ReverseProxy),
	}
}

func (p *Proxy) AddTarget(pathPrefix, targetURL string) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Customize the director to strip the prefix if needed
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Strip the /api/v1/prefix if necessary
		// For this project, we might want to keep it or strip it depending on how services are implemented.
		// Usually, internal services don't expect the /api/v1 prefix.
		if strings.HasPrefix(req.URL.Path, "/api/v1") {
			// e.g. /api/v1/users/me -> /users/me
			// But the spec says: /api/v1/users/* -> user-service:8081
			// And user-service has r.Route("/users", ...)
			// So /api/v1/users/me should probably become /users/me
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		}
	}

	p.targets[pathPrefix] = proxy
	return nil
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	for prefix, proxy := range p.targets {
		if strings.HasPrefix(r.URL.Path, prefix) {
			proxy.ServeHTTP(w, r)
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}
