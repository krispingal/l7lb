package usecases

import (
	"sync/atomic"

	"github.com/krispingal/l7lb/internal/domain"
)

type RoundRobinStrategy struct {
	current uint32
}

func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

func (rr *RoundRobinStrategy) GetNextBackend(backends []*domain.Backend) (*domain.Backend, error) {
	index := atomic.AddUint32(&rr.current, 1) - 1
	return backends[index%uint32(len(backends))], nil
}
