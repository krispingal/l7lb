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
	routingTable       *infrastructure.RoutingTable
	healthySet         sync.Map   // Map for lookups
	mu                 sync.Mutex // To protect healthySet during notifications
	httpClient         *http.Client
	logger             *zap.Logger
}

func NewHealthChecker(healthyFreq time.Duration, unhealthyFreq time.Duration, routingTable *infrastructure.RoutingTable, httpClient *http.Client, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		serverChan:         make(chan *domain.Backend, 1000),
		routingTable:       routingTable,
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
	// Assume new servers are initially added to the queue, all servers needs to be verified by health checker to
	hc.serverChan <- backend
}

// Worker for checking servers' health status
func (hc *HealthChecker) checkServers() {
	for backend := range hc.serverChan {
		checkFrequency := hc.healthyFrequency
		_, isNotHealthy := hc.healthySet.Load(backend.URL)
		if isNotHealthy {
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
func (hc *HealthChecker) updateBackendStatus(backend *domain.Backend, isHealthy bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	_, exists := hc.healthySet.Load(backend.URL)
	if isHealthy {
		if !exists {
			hc.healthySet.Store(backend.URL, backend)
			hc.logger.Info("Backend moved to healthy", zap.String("backend_url", backend.URL))
			hc.notifyLoadBalancer(backend, isHealthy)
		}
	} else {
		if !exists {
			hc.healthySet.Delete(backend.URL)
			hc.logger.Info("Backend moved to unhealthy", zap.String("backend_url", backend.URL))
			hc.notifyLoadBalancer(backend, isHealthy) // Notify immediately on change
		}
	}

}

// notifyLoadBalancer sends the updated healthy list to the load balancer
func (hc *HealthChecker) notifyLoadBalancer(backend *domain.Backend, isHealthy bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	listeners, exists := hc.routingTable.GetLoadBalancersForBackend(backend.URL)
	if !exists {
		hc.logger.Error("No load balancers found for backend", zap.String("backend_url", backend.URL))
		return
	}
	groupId, exists := hc.routingTable.GetBackendGroupOfBackend(backend.URL)
	healthUpdate := &domain.BackendHealthUpdate{
		Backend:   backend,
		IsHealthy: isHealthy,
		GroupId:   groupId,
	}
	// Only notify the relevant listeners/load balancers
	for _, listener := range listeners {
		go listener.ListenForHealthUpdates(healthUpdate)
	}
}
