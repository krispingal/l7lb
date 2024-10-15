package infrastructure

import (
	"sync"

	"github.com/krispingal/l7lb/internal/domain"
)

type RoutingTable struct {
	// Maps routes to backend group
	routeToBackendGroups        sync.Map     // map[string][]*domain.Backend
	backendToLoadBalancers      sync.Map     // map[*domain.Backend]*domain.LoadBalancer
	backendGroupToLoadBalancers sync.Map     // map[string][]*domain.LoadBalancer
	backendToGroup              sync.Map     // map[string][string]
	mu                          sync.RWMutex // Protects the maps during updates
}

// NewRoutingTable initializes a new thread-safe routing table
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		routeToBackendGroups:        sync.Map{},
		backendToLoadBalancers:      sync.Map{},
		backendGroupToLoadBalancers: sync.Map{},
		backendToGroup:              sync.Map{},
	}
}

// AddBackendGroup adds a backend group for a given group Id
func (rt *RoutingTable) AddBackendGroup(groupId string, backends []*domain.Backend, lbs []*domain.LoadBalancer) error {
	rt.routeToBackendGroups.Store(groupId, backends)
	rt.backendGroupToLoadBalancers.Store(groupId, lbs)
	for _, backend := range backends {
		rt.backendToGroup.Store(backend.URL, groupId)
		rt.backendToLoadBalancers.Store(backend.URL, lbs)
	}
	return nil
}

// GetBackendsForGroup retrieves the backends for a given backend group Id
func (rt *RoutingTable) GetBackendsForGroup(groupId string) ([]*domain.Backend, bool) {
	backends, ok := rt.routeToBackendGroups.Load(groupId)
	if !ok {
		return nil, false
	}
	return backends.([]*domain.Backend), true
}

// GetLoadBalancesrForGroup gets the load balancers for a backend group (used in routing decisions)
func (rt *RoutingTable) GetLoadBalancersForGroup(groupId string) ([]*domain.LoadBalancer, error) {
	lbs, ok := rt.backendGroupToLoadBalancers.Load(groupId)
	if !ok || len(lbs.([]*domain.LoadBalancer)) == 0 {
		return nil, nil
	}
	return lbs.([]*domain.LoadBalancer), nil
}

// GetLoadBalancersForBackend retrieves the load balancers for a specific backend (used for health updates)
func (rt *RoutingTable) GetLoadBalancersForBackend(backendURL string) ([]*domain.LoadBalancer, bool) {
	lbs, ok := rt.backendToLoadBalancers.Load(backendURL)
	if !ok {
		return nil, false
	}
	return lbs.([]*domain.LoadBalancer), true
}

// GetBackendGroupForBackend retrieves the group Id for a backend
func (rt *RoutingTable) GetBackendGroupOfBackend(backendURL string) (string, bool) {
	groupId, ok := rt.backendToGroup.Load(backendURL)
	if !ok {
		return "", false
	}
	return groupId.(string), true
}
