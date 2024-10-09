package httphandler

import (
	"net/http"
	"strings"

	"github.com/krispingal/l7lb/internal/usecases/loadbalancing"
)

func NewPathRouterWithLB(routes map[string]*loadbalancing.LoadBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, lb := range routes {
			if strings.HasPrefix(r.URL.Path, path) {
				lb.RouteRequest(w, r)
				return
			}
		}
	})
}

func NewPathRouterExactPathWithLB(routes map[string]*loadbalancing.LoadBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Normalize the path by trimming trailing slashes
		path := strings.TrimSuffix(r.URL.Path, "/")
		if lb, exists := routes[path]; exists {
			lb.RouteRequest(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
