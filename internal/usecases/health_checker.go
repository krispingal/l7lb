package usecases

import (
	"log"
	"net/http"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
)

type HealthChecker struct {
	backends   []*domain.Backend
	httpClient *http.Client
}

func NewHealthChecker(backends []*domain.Backend, httpClient *http.Client) *HealthChecker {
	return &HealthChecker{
		backends:   backends,
		httpClient: httpClient,
	}
}

func (hc *HealthChecker) Start() {
	for _, backend := range hc.backends {
		go func(b *domain.Backend) {
			for {
				resp, err := hc.httpClient.Get(b.URL + b.Health)
				if err != nil || resp.StatusCode != http.StatusOK {
					b.Alive = false
					log.Printf("Backend %s is down\n", b.URL)
				} else {
					b.Alive = true
					log.Printf("Backend %s is up\n", b.URL)
				}
				time.Sleep(10 * time.Second)
			}
		}(backend)
	}
}
