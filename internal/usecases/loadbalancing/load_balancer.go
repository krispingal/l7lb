package loadbalancing

import (
	"io"
	"net/http"
	"sync"

	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"go.uber.org/zap"
)

type LoadBalancer struct {
	backends []*domain.Backend
	strategy LoadBalancingStrategy
	logger   *zap.Logger
}

func NewLoadBalancer(backends []*domain.Backend, strategy LoadBalancingStrategy, logger *zap.Logger) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
		strategy: strategy,
		logger:   logger,
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
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: 10 * time.Second, // Set a timeout for the overall backend requests
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64<<10) // 64KB buffer
	},
}

func (lb *LoadBalancer) RouteRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	backend, err := lb.strategy.GetNextBackend(lb.backends)

	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		lb.logger.Error("Load balancer did not receive a next backend")
		return
	}

	if !backend.Alive {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		lb.logger.Error("Backend unavailable", zap.String("url", backend.URL), zap.Int("status", http.StatusServiceUnavailable), zap.Duration("duration", time.Since(startTime)))
		return
	}

	targetURL := backend.URL + r.URL.Path + "?" + r.URL.RawQuery

	// Create a new request
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Backend request error", http.StatusInternalServerError)
		lb.logger.Error("Backend unavailable", zap.String("url", backend.URL), zap.Int("status", http.StatusServiceUnavailable), zap.Duration("duration", time.Since(startTime)))
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
		if resp != nil && (resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429) {
			break
		}
		if err != nil || resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lb.logger.Error("Error making request to backend", zap.String("backend_url", backend.URL), zap.Error(err))
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
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	io.CopyBuffer(w, resp.Body, buf)
	lb.logger.Debug("Request routed successfully", zap.String("url", backend.URL), zap.Int("status", resp.StatusCode), zap.Duration("duration", time.Since(startTime)))
}
