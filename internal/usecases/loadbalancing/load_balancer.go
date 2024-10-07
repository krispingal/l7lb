package loadbalancing

import (
	"io"
	"log"
	"net/http"

	"time"

	"github.com/krispingal/l7lb/internal/domain"
)

type LoadBalancer struct {
	backends []*domain.Backend
	strategy LoadBalancingStrategy
}

func NewLoadBalancer(backends []*domain.Backend, strategy LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
		strategy: strategy,
	}
}

func (lb *LoadBalancer) Backends() []*domain.Backend {
	return lb.backends
}

func (lb *LoadBalancer) Strategy() LoadBalancingStrategy {
	return lb.strategy
}

func (lb *LoadBalancer) RouteRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	backend, err := lb.strategy.GetNextBackend(lb.backends)

	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		log.Printf("Load balancer did not receive a next backend")
		return
	}

	if !backend.Alive {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		log.Printf("Backend: %s | Status: %d | Latency: %v\n", backend.URL, http.StatusServiceUnavailable, time.Since(startTime))
		return
	}

	targetURL := backend.URL + r.URL.Path + "?" + r.URL.RawQuery
	resp, err := http.Get(targetURL)
	if err != nil || resp.StatusCode >= 500 {
		http.Error(w, "Backend unavailable", http.StatusServiceUnavailable)
		log.Printf("Backend: %s | Status: %d | Latency: %v\n", backend.URL, http.StatusServiceUnavailable, time.Since(startTime))
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	log.Printf("Backend: %s | Status: %d | Latency: %v\n", backend.URL, resp.StatusCode, time.Since(startTime))
}
