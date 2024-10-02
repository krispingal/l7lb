package httphandler

import (
	"net/http"
	"strings"

	"github.com/krispingal/l7lb/internal/usecases"
)

func NewPathRouterWithLB(routes map[string]*usecases.LoadBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, lb := range routes {
			if strings.HasPrefix(r.URL.Path, path) {
				lb.Forward(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
}
