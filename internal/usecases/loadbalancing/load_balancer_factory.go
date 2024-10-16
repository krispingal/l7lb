package loadbalancing

import (
	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/infrastructure"
	"github.com/krispingal/l7lb/internal/usecases"
	"go.uber.org/zap"
)

func CreateLoadBalancers(config *infrastructure.Config, registry *infrastructure.BackendRegistry, healthchecker *usecases.HealthChecker, logger *zap.Logger) map[string]*LoadBalancer {
	lbMap := make(map[string]*LoadBalancer)

	for _, route := range config.Routes {
		builder := NewLoadBalancerBuilder()
		var healthUpdateChannels []<-chan domain.BackendStatus
		for _, backend := range route.Backends {
			channel := registry.Subscribe(backend.URL)
			healthUpdateChannels = append(healthUpdateChannels, channel)
		}
		builder.WithHealthUpdateChannels(healthUpdateChannels)
		builder.WithBackendRegistry(registry)

		strategy := NewRoundRobinStrategy()
		builder.WithStrategy(strategy)
		builder.WithLogger(logger)

		lbMap[route.Path] = builder.Build()
		for _, configBackend := range route.Backends {
			backend := domain.NewBackend(configBackend.URL, configBackend.Health)
			healthchecker.AddBackend(backend)
			registry.AddBackendToRegistry(*backend)
		}
	}
	logger.Debug("Created load balancers")
	return lbMap
}
