package usecases

import (
	"log"
	"net/http"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
)

type HealthChecker struct {
	healthyServers     chan *domain.Backend
	unhealthyServers   chan *domain.Backend
	healthyFrequency   time.Duration
	unhealthyFrequency time.Duration
	httpClient         *http.Client
}

func NewHealthChecker(healthyFreq time.Duration, unhealthyFreq time.Duration, httpClient *http.Client) *HealthChecker {
	return &HealthChecker{
		healthyServers:     make(chan *domain.Backend, 1000),
		unhealthyServers:   make(chan *domain.Backend, 1000),
		healthyFrequency:   healthyFreq,
		unhealthyFrequency: unhealthyFreq,
		httpClient:         httpClient,
	}
}

// Worker for checking healthy servers
func (hc *HealthChecker) checkHealthyServers() {
	for backend := range hc.healthyServers {
		resp, err := hc.httpClient.Get(backend.URL + backend.Health)
		if err != nil || resp.StatusCode != http.StatusOK {
			backend.Alive = false
			log.Printf("Backend %s moved to unhealthy\n", backend.URL)
			hc.unhealthyServers <- backend
		}
		time.Sleep(hc.healthyFrequency)
	}
}

// Worker for checking unhealthy & new servers
func (hc *HealthChecker) checkUnhealthyServers() {
	for backend := range hc.unhealthyServers {
		resp, err := hc.httpClient.Get(backend.URL + backend.Health)
		if err == nil && resp.StatusCode == http.StatusOK {
			backend.Alive = true
			log.Printf("Backend %s moved to healthy\n", backend.URL)
			hc.healthyServers <- backend
		}
		time.Sleep(hc.unhealthyFrequency)
	}
}

// Start launches the health check workers
func (hc *HealthChecker) Start() {
	go hc.checkHealthyServers()
	go hc.checkUnhealthyServers()
}

// AddBackend adds a backend to the unhealthyServers channel initially
func (hc *HealthChecker) AddBackend(backend *domain.Backend) {
	// Assume new servers are initially added to the unhealthy queue
	hc.unhealthyServers <- backend
}
