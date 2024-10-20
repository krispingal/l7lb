package loadbalancing

import (
	"io"
	"net/http"
	"reflect"
	"sync"

	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/infrastructure"
	"go.uber.org/zap"
)

type LoadBalancer struct {
	backendRegistry      *infrastructure.BackendRegistry
	strategy             LoadBalancingStrategy
	logger               *zap.Logger
	healthUpdateChannels []<-chan domain.BackendStatus
	healthyBackends      []*domain.Backend
	mu                   sync.RWMutex // mutex to protect healthy backend list
}

func NewLoadBalancer(registry *infrastructure.BackendRegistry, strategy LoadBalancingStrategy, healthChannels []<-chan domain.BackendStatus, logger *zap.Logger) *LoadBalancer {
	lb := &LoadBalancer{
		backendRegistry:      registry,
		strategy:             strategy,
		healthUpdateChannels: healthChannels,
		logger:               logger,
	}

	go lb.listenToHealthUpdates()
	return lb
}

func (lb *LoadBalancer) Strategy() LoadBalancingStrategy {
	return lb.strategy
}

func (lb *LoadBalancer) listenToHealthUpdates() {
	cases := make([]reflect.SelectCase, len(lb.healthUpdateChannels))
	lb.logger.Info("Listening for health updates in loadbalancer")

	for i, ch := range lb.healthUpdateChannels {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
	}

	for {
		// Wait for any of the channels to receive a value
		chosen, value, ok := reflect.Select(cases)
		if ok {
			update := value.Interface().(domain.BackendStatus)
			lb.logger.Debug("Received backend health update", zap.String("backend_url", update.URL))
			lb.updateProcessDispatcher(update)
		} else {
			lb.logger.Warn("BackendHealthUpdateChannel was closed", zap.Int("update_channel", chosen))
		}
	}
}

func (lb *LoadBalancer) updateProcessDispatcher(update domain.BackendStatus) {
	if update.IsHealthy {
		lb.addToHealthyBackends(update.URL)
	} else {
		lb.removeFromHealthyBackends(update.URL)
	}
}

func (lb *LoadBalancer) addToHealthyBackends(backendURL string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, backend := range lb.healthyBackends {
		if backend.URL == backendURL {
			return
		}
	}
	backend, ok := lb.backendRegistry.GetBackendByURL(backendURL)
	if !ok {
		lb.logger.Error("No backend forund for url", zap.String("backend_url", backendURL))
		return
	}
	lb.healthyBackends = append(lb.healthyBackends, &backend)
}

func (lb *LoadBalancer) removeFromHealthyBackends(backendURL string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, backend := range lb.healthyBackends {
		if backend.URL == backendURL {
			lb.healthyBackends = append(lb.healthyBackends[:i], lb.healthyBackends[i+1:]...)
			return
		}
	}
}

func (lb *LoadBalancer) getHealthyBackends() []*domain.Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.healthyBackends
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
	backends := lb.getHealthyBackends()
	if len(backends) == 0 {
		lb.logger.Error("No healthy backends available", zap.Any("request_url", r.URL))
	}
	backend, err := lb.strategy.GetNextBackend(backends)

	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		lb.logger.Error("Load balancer did not receive a next backend")
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
	lb.logger.Debug("Request routed successfully", zap.String("backend_url", backend.URL), zap.Int("status", resp.StatusCode), zap.Duration("duration", time.Since(startTime)))
}
