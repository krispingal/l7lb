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

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: 10 * time.Second, // Set a timeout for the overall backend requests
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

	// Create a new request
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Backend request error", http.StatusInternalServerError)
		log.Printf("Backend: %s | Status: %d | Latency: %v | error: %v\n", backend.URL, http.StatusServiceUnavailable, time.Since(startTime), err)
		return
	}
	defer req.Body.Close()

	req.Header = r.Header // Pass the original header
	maxRetries := 3
	var resp *http.Response
	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			break
		}
		// It is a Client-side (4xx) - do not retry
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
		if err != nil || resp.StatusCode >= 500 {
			log.Printf("Error making request to backend %s: %v", backend.URL, err)
			time.Sleep(time.Duration(i) * time.Second) // Exponential backoff
		} else {
			// For other errors - non transient, break the loop
			break
		}
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	buf := make([]byte, 32<<10) // Use a 32KB buffer
	io.CopyBuffer(w, resp.Body, buf)
	log.Printf("Backend: %s | Status: %d | Latency: %v\n", backend.URL, resp.StatusCode, time.Since(startTime))
}
