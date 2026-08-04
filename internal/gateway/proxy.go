package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Route struct {
	Prefix         string
	Target         string
	StripAPIPrefix bool
}

type Proxy struct {
	routes []compiledRoute
}

type compiledRoute struct {
	prefix         string
	stripAPIPrefix bool
	proxy          *httputil.ReverseProxy
}

func NewProxy(routes []Route) (*Proxy, error) {
	compiled := make([]compiledRoute, 0, len(routes))
	for _, route := range routes {
		target, err := url.Parse(strings.TrimRight(route.Target, "/"))
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		originalDirector := proxy.Director
		stripAPIPrefix := route.StripAPIPrefix
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			if stripAPIPrefix && strings.HasPrefix(req.URL.Path, "/api/") {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
			}
			req.Host = target.Host
		}
		compiled = append(compiled, compiledRoute{
			prefix:         route.Prefix,
			stripAPIPrefix: stripAPIPrefix,
			proxy:          proxy,
		})
	}
	return &Proxy{routes: compiled}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health/"+ServiceName || r.URL.Path == "/actuator/health" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"UP","service":"api-gateway"}`))
		return
	}
	for _, route := range p.routes {
		if strings.HasPrefix(r.URL.Path, route.prefix) {
			route.proxy.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}
