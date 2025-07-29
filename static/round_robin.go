package static

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/DucTran999/load-balancing-algo/backend"
	"github.com/DucTran999/load-balancing-algo/errs"
)

type roundRobin struct {
	backends   []*backend.SimpleHTTPServer
	counter    uint64
	proxyCache sync.Map
}

func NewRoundRobinAlg(targets []*backend.SimpleHTTPServer) (*roundRobin, error) {
	if len(targets) == 0 {
		return nil, errs.ErrNoTargetServersFound
	}

	// Validate backend URLs
	for _, target := range targets {
		if target.GetUrl() == nil {
			return nil, errs.ErrInvalidBackendUrl
		}
	}

	return &roundRobin{
		backends:   targets,
		proxyCache: sync.Map{},
	}, nil
}

func (lb *roundRobin) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	nextUrl := lb.getNextBackend()

	// Log the next URL to which the request will be forwarded
	log.Printf("[INFO] load balancer forwarding request to: %v\n", nextUrl.String())

	// Create a reverse proxy for the next backend
	proxy := lb.getOrCreateProxy(&nextUrl)

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func (lb *roundRobin) getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.String()
	if proxy, ok := lb.proxyCache.Load(key); ok {
		return proxy.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	lb.proxyCache.Store(key, proxy)

	return proxy
}

func (lb *roundRobin) getNextBackend() url.URL {
	idx := atomic.AddUint64(&lb.counter, 1)

	next := lb.backends[idx%uint64(len(lb.backends))]

	return *next.GetUrl()
}
