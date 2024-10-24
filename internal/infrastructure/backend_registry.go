package infrastructure

import (
	"log"
	"sync"

	"github.com/krispingal/l7lb/internal/domain"
)

type healthUpdateChannel chan domain.BackendStatus

type BackendRegistry struct {
	mu          sync.RWMutex
	backendId   map[uint64]domain.Backend        //backendId -> domain.Backend
	backends    map[uint64]domain.BackendStatus  // Store backend health status
	subscribers map[uint64][]healthUpdateChannel // Map of backend -> list of load balancer channel
}

// Initialize the registry
func NewBackendRegistry() *BackendRegistry {
	return &BackendRegistry{
		backendId:   make(map[uint64]domain.Backend),
		backends:    make(map[uint64]domain.BackendStatus),
		subscribers: make(map[uint64][]healthUpdateChannel),
	}
}

// Method to update the health status of a backend
func (r *BackendRegistry) UpdateHealth(status domain.BackendStatus) error {
	if status == (domain.BackendStatus{}) {
		log.Fatal("status is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.backends[status.Id] = status

	// Notify all lbs of this backend health update
	if subs, ok := r.subscribers[status.Id]; ok {
		for _, ch := range subs {
			ch <- status // Non-blocking send
		}
	}
	return nil
}

// Method for load balancers to subscribe to specific backend health updates
func (r *BackendRegistry) Subscribe(backendId uint64) <-chan domain.BackendStatus {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a channel for updates
	ch := make(chan domain.BackendStatus, 10)

	// Add the channel to the list of subscribers for the backend
	r.subscribers[backendId] = append(r.subscribers[backendId], ch)

	return ch
}

// Get the backend with id
func (r *BackendRegistry) GetBackendById(backendId uint64) (domain.Backend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	backend, exists := r.backendId[backendId]
	return backend, exists
}

func (r *BackendRegistry) AddBackendToRegistry(backend domain.Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backendId[backend.Id] = backend
}
