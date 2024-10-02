package usecases

import (
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
)

type LoadBalancer struct {
	backends []*domain.Backend
	current  uint32
}

func NewLoadBalancer(backends []*domain.Backend) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
	}
}

func (lb *LoadBalancer) getNextBackend() *domain.Backend {
	index := atomic.AddUint32(&lb.current, 1) % uint32(len(lb.backends))
	return lb.backends[index]
}

func (lb *LoadBalancer) Forward(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	backend := lb.getNextBackend()

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

func (lb *LoadBalancer) StartHealthCheck() {
	for _, backend := range lb.backends {
		go func(b *domain.Backend) {
			for {
				resp, err := http.Get(b.URL + b.Health)
				if err != nil || resp.StatusCode != http.StatusOK {
					b.Alive = false
					log.Printf("Backend %s is down\n", b.URL)
				} else {
					b.Alive = true
					log.Printf("Backend %s is up\n", b.URL)
				}
				time.Sleep(10 * time.Second)
			}
		}(backend)
	}
}
