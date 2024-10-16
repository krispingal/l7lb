package loadbalancing

import (
	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/infrastructure"
	"go.uber.org/zap"
)

type LoadBalancerBuilder struct {
	backends       []*domain.Backend
	registry       *infrastructure.BackendRegistry
	updateChannels []<-chan domain.BackendStatus
	strategy       LoadBalancingStrategy
	logger         *zap.Logger
}

// NewLoadBalancerBuilder initializes the builder
func NewLoadBalancerBuilder() *LoadBalancerBuilder {
	return &LoadBalancerBuilder{}
}

// WithStrategy sets the load balancing strategy
func (b *LoadBalancerBuilder) WithStrategy(strategy LoadBalancingStrategy) *LoadBalancerBuilder {
	b.strategy = strategy
	return b
}

// WithHealthUpdateChannels sets the health update channel
func (b *LoadBalancerBuilder) WithHealthUpdateChannels(updateChannels []<-chan domain.BackendStatus) *LoadBalancerBuilder {
	b.updateChannels = updateChannels
	return b
}

// WithHealthUpdateChannels sets the health update channel
func (b *LoadBalancerBuilder) WithBackendRegistry(registry *infrastructure.BackendRegistry) *LoadBalancerBuilder {
	b.registry = registry
	return b
}

// WithLogger sets the logger
func (b *LoadBalancerBuilder) WithLogger(logger *zap.Logger) *LoadBalancerBuilder {
	b.logger = logger
	return b
}

// Build creates the final LoadBalancer object
func (b *LoadBalancerBuilder) Build() *LoadBalancer {
	return NewLoadBalancer(b.registry, b.strategy, b.updateChannels, b.logger)
}
