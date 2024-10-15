package infrastructure

import (
	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/infrastructure"
	"github.com/krispingal/l7lb/internal/usecases/loadbalancing"
)

type RouteManager struct {
	RoutingTable *infrastructure.RoutingTable
	RouteMapping map[string]string // Maps route paths to backend group Ids
}

func NewRouteManager(config Config, logger *zap.Logger) *RouteManager {
	routingTable := NewRoutingTable()
	routeMapping := make(map[string]string)

	for _, group := range config.BackendGroups {
		var backends []*domain.Backend
		for _, backendConfig := range group.Backends {
			backend := domain.NewBackend(backendConfig.URL, backendConfig.Health)
			backends = append(backends, backend)
		}
		builder := loadbalancing.NewLoadBalancerBuilder()
		strategy := loadbalancing.NewRoundRobinStrategy()
		builder.WithStrategy(strategy)
		lb := builder.Build()
		routingTable.AddBackendGroup(group.GroupId, backends, lb)
	}
	for _, route := range config.Routes {
		routeMapping[route.Path] = route.GroupId
	}
	return &RouteManager{RoutingTable: routingTable, RouteMapping: routeMapping}
}

// GetLoadBalancerAndGroupForRoute returns the load balancer and the the backend group for a route path
func (rm *RouteManager) GetLoadBalancersAndGroupForRoute(routePath string) ([]*domain.LoadBalancer, string, error) {
	if groupId, ok := rm.RouteMapping[routePath]; !ok {
		return nil, "", nil // TODO add error here
	}
	
	if lbs, ok := rm.RoutingTable.GetLoadBalancersForGroup(groupId); !ok {
		return nil, groupId, nil
	}
	return lbs, groupId, nil
}
