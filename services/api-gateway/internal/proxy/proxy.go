package proxy

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
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
	proxy.FlushInterval = -1 // Flush immediately for SSE/streaming

	// Configure custom transport for high concurrency
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.MaxIdleConns = 1000
	customTransport.MaxIdleConnsPerHost = 1000
	customTransport.MaxConnsPerHost = 1000
	customTransport.IdleConnTimeout = 90 * time.Second
	proxy.Transport = customTransport

	// Customize the director to strip the prefix if needed
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Set the host header to the target host
		req.Host = target.Host

		// Ensure WebSocket headers are preserved and correctly set
		if strings.ToLower(req.Header.Get("Upgrade")) == "websocket" {
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
		}

		// Propagate Request ID
		if reqID := req.Header.Get("X-Request-ID"); reqID != "" {
			req.Header.Set("X-Request-ID", reqID)
		}

		// Strip the /api/v1 prefix
		if strings.HasPrefix(req.URL.Path, "/api/v1") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
		}
	}

	p.targets[pathPrefix] = proxy
	return nil
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Find the longest matching prefix for more accurate routing
	var bestPrefix string
	var bestProxy *httputil.ReverseProxy

	for prefix, proxy := range p.targets {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > len(bestPrefix) {
				bestPrefix = prefix
				bestProxy = proxy
			}
		}
	}

	if bestProxy != nil {
		log.Info().Str("path", path).Str("prefix", bestPrefix).Msg("Proxying request")
		bestProxy.ServeHTTP(w, r)
		return
	}

	log.Warn().Str("path", path).Msg("No proxy target found")
	http.Error(w, "not found", http.StatusNotFound)
}
