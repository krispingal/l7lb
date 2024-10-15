package loadbalancing

import (
		"github.com/krispingal/l7lb/internal/domain"
	"go.uber.org/zap"
)

type LoadBalancerBuilder struct {
	strategy LoadBalancingStrategy
	logger   *zap.Logger
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

// WithLogger sets the logger
func (b *LoadBalancerBuilder) WithLogger(logger *zap.Logger) *LoadBalancerBuilder {
	b.logger = logger
	return b
}

// Build creates the final LoadBalancer object
func (b *LoadBalancerBuilder) Build() domain.LoadBalancer {
	return NewLoadBalancer(b.strategy, b.logger)
}
