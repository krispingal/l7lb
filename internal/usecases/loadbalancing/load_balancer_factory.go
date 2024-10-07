package loadbalancing

import (
	config "github.com/krispingal/l7lb/internal/infrastructure"
	"github.com/krispingal/l7lb/internal/usecases"
)

func CreateLoadBalanacers(config *config.Config) map[string]*LoadBalancer {
	lbMap := make(map[string]*LoadBalancer)

	for _, route := range config.Routes {
		builder := NewLoadBalancerBuilder()

		builder.WithBackends(route.Backends)

		strategy := usecases.NewRoundRobinStrategy()
		builder.WithStrategy(strategy)

		lbMap[route.Path] = builder.Build()
	}
	return lbMap
}
