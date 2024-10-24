package loadbalancing

import (
	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/infrastructure"
	"github.com/krispingal/l7lb/internal/usecases"
	"go.uber.org/zap"
)

func CreateLoadBalancers(config *infrastructure.Config, registry *infrastructure.BackendRegistry, healthChecker *usecases.HealthChecker, logger *zap.Logger) map[string]*LoadBalancer {
	lbMap := make(map[string]*LoadBalancer)

	for _, route := range config.Routes {
		healthUpdateChannels := setupHealthAndRegister(route.Backends, registry, healthChecker)
		builder := NewLoadBalancerBuilder().
			WithBackendRegistry(registry).
			WithStrategy(NewRoundRobinStrategy()).
			WithHealthUpdateChannels(healthUpdateChannels).
			WithLogger(logger)

		lbMap[route.Path] = builder.Build()
	}
	logger.Debug("Created load balancers")
	return lbMap
}

func setupHealthAndRegister(backends []infrastructure.Backend, registry *infrastructure.BackendRegistry, healthChecker *usecases.HealthChecker) []<-chan domain.BackendStatus {
	var healthUpdateChannels []<-chan domain.BackendStatus
	for _, backendConfig := range backends {
		// Register the backend with health checker and registry
		backend := registerBackend(backendConfig, registry, healthChecker)
		// Subscribe for health updates
		channel := registry.Subscribe(backend.Id)
		healthUpdateChannels = append(healthUpdateChannels, channel)
	}
	return healthUpdateChannels
}

func registerBackend(backendConfig infrastructure.Backend, registry *infrastructure.BackendRegistry, healthChecker *usecases.HealthChecker) *domain.Backend {
	backend := domain.NewBackend(backendConfig.URL, backendConfig.Health)
	healthChecker.AddBackend(backend)
	registry.AddBackendToRegistry(*backend)
	return backend
}
