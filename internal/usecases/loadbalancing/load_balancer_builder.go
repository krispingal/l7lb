package loadbalancing

import (
	"github.com/krispingal/l7lb/internal/domain"
	config "github.com/krispingal/l7lb/internal/infrastructure"
	"go.uber.org/zap"
)

type LoadBalancerBuilder struct {
	backends []*domain.Backend
	strategy LoadBalancingStrategy
	logger   *zap.Logger
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

// WithLogger sets the logger
func (b *LoadBalancerBuilder) WithLogger(logger *zap.Logger) *LoadBalancerBuilder {
	b.logger = logger
	return b
}

// Build creates the final LoadBalancer object
func (b *LoadBalancerBuilder) Build() *LoadBalancer {
	return NewLoadBalancer(b.backends, b.strategy, b.logger)
}
