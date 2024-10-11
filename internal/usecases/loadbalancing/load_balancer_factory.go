package loadbalancing

import (
	config "github.com/krispingal/l7lb/internal/infrastructure"
	"go.uber.org/zap"
)

func CreateLoadBalancers(config *config.Config, logger *zap.Logger) map[string]*LoadBalancer {
	lbMap := make(map[string]*LoadBalancer)

	for _, route := range config.Routes {
		builder := NewLoadBalancerBuilder()

		builder.WithBackends(route.Backends)

		strategy := NewRoundRobinStrategy()
		builder.WithStrategy(strategy)
		builder.WithLogger(logger)

		lbMap[route.Path] = builder.Build()
	}
	return lbMap
}
