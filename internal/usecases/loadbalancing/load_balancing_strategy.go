package loadbalancing

import "github.com/krispingal/l7lb/internal/domain"

type LoadBalancingStrategy interface {
	GetNextBackend([]*domain.Backend) (*domain.Backend, error)
}
