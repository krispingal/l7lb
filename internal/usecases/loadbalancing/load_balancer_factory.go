package loadbalancing

// import (
// 	"github.com/krispingal/l7lb/internal/infrastructure"
// 	"go.uber.org/zap"
// )

// func CreateLoadBalancers(logger *zap.Logger) map[string]*LoadBalancer {
// 	lbMap := make(map[string]*LoadBalancer)

// 	for route, backends := range routeManager.Routes {
// 		builder := NewLoadBalancerBuilder()
// 		strategy := NewRoundRobinStrategy()
// 		builder.WithStrategy(strategy)
// 		builder.WithLogger(logger)

// 		lb := builder.Build()
// 		for _, backend := range backends {
// 			healthChecker.AddBackend(backend)
// 		}
// 	}
// 	return lbMap
// }
