package usecases

import (
	"net/http"
	"sync"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"go.uber.org/zap"
)

type HealthChecker struct {
	serverChan         chan *domain.Backend
	HealthyBackendChan chan []*domain.Backend // Channel to push updated list of healthy backends
	healthyFrequency   time.Duration
	unhealthyFrequency time.Duration
	healthySet         sync.Map   // Map for lookups
	mu                 sync.Mutex // To protect healthySet during notifications
	httpClient         *http.Client
	logger             *zap.Logger
}

func NewHealthChecker(healthyFreq time.Duration, unhealthyFreq time.Duration, httpClient *http.Client, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		serverChan:         make(chan *domain.Backend, 1000),
		HealthyBackendChan: make(chan []*domain.Backend, 10),
		healthyFrequency:   healthyFreq,
		unhealthyFrequency: unhealthyFreq,
		httpClient:         httpClient,
		logger:             logger,
	}
}

// Start launches the health check workers
func (hc *HealthChecker) Start() {
	for i := 0; i < 2; i++ {
		go hc.checkServers()
	}
}

// AddBackend adds a backend to the unhealthyServers channel initially
func (hc *HealthChecker) AddBackend(backend *domain.Backend) {
	// Assume new servers are initially added to the unhealthy queue
	backend.Alive = false
	hc.serverChan <- backend
}

// Worker for checking servers' health status
func (hc *HealthChecker) checkServers() {
	for backend := range hc.serverChan {
		checkFrequency := hc.healthyFrequency
		if !backend.Alive {
			checkFrequency = hc.unhealthyFrequency
		}
		// Pereform health check
		resp, err := hc.httpClient.Get(backend.URL + backend.Health)
		if err == nil && resp.StatusCode == http.StatusOK {
			hc.updateBackendStatus(backend, true) // Mark as healthy
		} else {
			hc.updateBackendStatus(backend, false) // Mark as unhealthy
		}
		time.Sleep(checkFrequency)
	}
}

// Updates Backend status and moves backend to approriate
func (hc *HealthChecker) updateBackendStatus(backend *domain.Backend, alive bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	_, exists := hc.healthySet.Load(backend.URL)
	if alive {
		backend.Alive = true
		if !exists {
			hc.healthySet.Store(backend.URL, backend)
			hc.logger.Info("Backend moved to healthy", zap.String("backend_url", backend.URL))
			hc.notifyLoadBalancer()
		}
	} else {
		backend.Alive = false
		if !exists {
			hc.healthySet.Delete(backend.URL)
			hc.logger.Info("Backend moved to unhealthy", zap.String("backend_url", backend.URL))
			hc.notifyLoadBalancer() // Notify immediately on change
		}
	}

}

// notifyLoadBalancer sends the updated healthy list to the load balancer
func (hc *HealthChecker) notifyLoadBalancer() {
	var healthyList []*domain.Backend

	hc.healthySet.Range(func(_, value interface{}) bool {
		healthyList = append(healthyList, value.(*domain.Backend))
		return true
	})

	// Send the healthy list to the load balancer
	hc.HealthyBackendChan <- healthyList
	hc.logger.Debug("Healthy backends list updated", zap.Int("count", len(healthyList)))
}
