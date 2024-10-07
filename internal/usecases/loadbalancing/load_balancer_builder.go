package loadbalancing

import (
	"github.com/krispingal/l7lb/internal/domain"
	config "github.com/krispingal/l7lb/internal/infrastructure"
)

type LoadBalancerBuilder struct {
	backends []*domain.Backend
	strategy LoadBalancingStrategy
}

// NewLoadBalancerBuilder initializes the builder
func NewLoadBalancerBuilder() *LoadBalancerBuilder {
	return &LoadBalancerBuilder{}
}

func (b *LoadBalancerBuilder) WithBackends(backendConfigs []config.Backend) *LoadBalancerBuilder {
	for _, backend := range backendConfigs {
		b.backends = append(b.backends, &domain.Backend{
			URL:    backend.URL,
			Alive:  true, // Assume that backend is alive initially
			Health: backend.Health,
		})
	}
	return b
}

// WithStrategy sets the load balancing strategy
func (b *LoadBalancerBuilder) WithStrategy(strategy LoadBalancingStrategy) *LoadBalancerBuilder {
	b.strategy = strategy
	return b
}

// Build creates the final LoadBAlancer object
func (b *LoadBalancerBuilder) Build() *LoadBalancer {
	return NewLoadBalancer(b.backends, b.strategy)
}
