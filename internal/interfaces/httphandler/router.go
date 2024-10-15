package httphandler

import (
	"net/http"
	"strings"

	"github.com/krispingal/l7lb/internal/infrastructure"
)

func NewPathRouterExactPathWithLB(routeManager *infrastructure.RouteManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Normalize the path by trimming trailing slashes
		path := strings.TrimSuffix(r.URL.Path, "/")
		if lbs, backendGroup, err := routeManager.GetLoadBalancersAndGroupForRoute(path); err != nil {
			http.NotFound(w, r)
		} else {
			// TODO use consistent hashing or other techniques to pick a loadbalancer
			lbs[0].RouteRequest(w, r, backendGroup)
		}
	})
}
