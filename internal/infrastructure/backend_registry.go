package infrastructure

import (
	"log"
	"sync"

	"github.com/krispingal/l7lb/internal/domain"
)

type healthUpdateChannel chan domain.BackendStatus

type BackendRegistry struct {
	mu          sync.RWMutex
	backendUrl  map[string]domain.Backend        // backendUrl -> domain.Backend
	backends    map[string]domain.BackendStatus  // Store backend health status
	subscribers map[string][]healthUpdateChannel // Map of backend -> list of load balancer channel
}

// Initialize the registry
func NewBackendRegistry() *BackendRegistry {
	return &BackendRegistry{
		backendUrl:  make(map[string]domain.Backend),
		backends:    make(map[string]domain.BackendStatus),
		subscribers: make(map[string][]healthUpdateChannel),
	}
}

// Method to update the health status of a backend
func (r *BackendRegistry) UpdateHealth(status domain.BackendStatus) error {
	if status == (domain.BackendStatus{}) {
		log.Fatal("status is empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.backends[status.URL] = status

	// Notify all lbs of this backend health update
	if subs, ok := r.subscribers[status.URL]; ok {
		for _, ch := range subs {
			ch <- status // Non-blocking send
		}
	}
	return nil
}

// Method for load balancers to subscribe to specific backend health updates
func (r *BackendRegistry) Subscribe(backendUrl string) <-chan domain.BackendStatus {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a channel for updates
	ch := make(chan domain.BackendStatus, 10)

	// Add the channel to the list of subscribers for the backend
	r.subscribers[backendUrl] = append(r.subscribers[backendUrl], ch)

	return ch
}

func (r *BackendRegistry) GetBackendByURL(backendUrl string) (domain.Backend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	backend, exists := r.backendUrl[backendUrl]
	return backend, exists
}

func (r *BackendRegistry) AddBackendToRegistry(backend domain.Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backendUrl[backend.URL] = backend
}
