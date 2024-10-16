package usecases

import (
	"net/http"
	"sync"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/infrastructure"
	"go.uber.org/zap"
)

type HealthChecker struct {
	serverChan         chan *domain.Backend
	healthyFrequency   time.Duration
	unhealthyFrequency time.Duration
	registry           *infrastructure.BackendRegistry
	healthySet         sync.Map   // Map for lookups
	mu                 sync.Mutex // To protect healthySet during notifications
	httpClient         *http.Client
	logger             *zap.Logger
}

func NewHealthChecker(healthyFreq time.Duration, unhealthyFreq time.Duration, registry *infrastructure.BackendRegistry, httpClient *http.Client, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		serverChan:         make(chan *domain.Backend, 1000),
		healthyFrequency:   healthyFreq,
		unhealthyFrequency: unhealthyFreq,
		registry:           registry,
		httpClient:         httpClient,
		logger:             logger,
	}
}

// Start launches the health check workers
func (hc *HealthChecker) Start() {
	numWorkers := 3
	for i := 0; i < numWorkers; i++ {
		go hc.worker(i)
	}
}

// AddBackend adds a backend to the unhealthyServers channel initially
func (hc *HealthChecker) AddBackend(backend *domain.Backend) {
	// Assume new servers are initially added to the unhealthy queue
	hc.logger.Debug("Added backend", zap.String("backend_url", backend.URL))
	backend.Alive = false
	hc.serverChan <- backend
}

func (hc *HealthChecker) worker(id int) {
	hc.logger.Info("Starting worker", zap.Int("worker_id", id))
	for backend := range hc.serverChan {
		hc.checkBackend(backend)
	}
}

// Worker for checking servers' health status
func (hc *HealthChecker) checkBackend(backend *domain.Backend) {
	checkFrequency := hc.healthyFrequency
	if !backend.Alive {
		checkFrequency = hc.unhealthyFrequency
	}
	// Pereform health check
	resp, err := hc.httpClient.Get(backend.URL + backend.Health)
	if err == nil && resp.StatusCode == http.StatusOK {
		hc.logger.Debug("Backend responded healthy", zap.String("backend_url", backend.URL))
		hc.updateBackendStatus(backend, true) // Mark as healthy
	} else {
		hc.logger.Debug("Backend responded not healthy", zap.String("backend_url", backend.URL))
		hc.updateBackendStatus(backend, false) // Mark as unhealthy
	}
	time.Sleep(checkFrequency)
	hc.serverChan <- backend
}

// Updates Backend status and moves backend to approriate
func (hc *HealthChecker) updateBackendStatus(backend *domain.Backend, isHealthy bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	_, exists := hc.healthySet.Load(backend.URL)
	if isHealthy {
		backend.Alive = true
		if !exists {
			hc.healthySet.Store(backend.URL, backend)
			statusUpdate := &domain.BackendStatus{URL: backend.URL, IsHealthy: isHealthy}
			hc.registry.UpdateHealth(*statusUpdate) // Notify immediately on change
			hc.logger.Info("Backend moved to healthy", zap.String("backend_url", backend.URL))
		}
	} else {
		backend.Alive = false
		if !exists {
			hc.healthySet.Delete(backend.URL)
			statusUpdate := &domain.BackendStatus{URL: backend.URL, IsHealthy: isHealthy}
			hc.registry.UpdateHealth(*statusUpdate) // Notify immediately on change
			hc.logger.Info("Backend moved to unhealthy", zap.String("backend_url", backend.URL))
		}
	}
}
