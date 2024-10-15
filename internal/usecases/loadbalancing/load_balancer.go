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
	strategy         LoadBalancingStrategy
	healthUpdateChan chan *domain.BackendHealthUpdate
	healthyBackends  sync.Map     // map[string][]*domain.Backend
	mu               sync.RWMutex // mutex to protect healthy backend list
	logger           *zap.Logger
}

func NewLoadBalancer(strategy LoadBalancingStrategy, logger *zap.Logger) domain.LoadBalancer {
	return &LoadBalancer{
		strategy:         strategy,
		healthUpdateChan: make(chan *domain.BackendHealthUpdate, 20),
		healthyBackends:  sync.Map{},
		logger:           logger,
	}
}

// HealthyBackends returns a list of healthy backends.
func (lb *LoadBalancer) HealthyBackends() []*domain.Backend {
	var allBackends []*domain.Backend
	lb.healthyBackends.Range(func(groupId, backend interface{}) bool {
		backends, ok := backend.([]*domain.Backend)
		if !ok {
			lb.logger.Error("Unable to construct backends")
			return false
		}
		allBackends = append(allBackends, backends...)
		return true
	})
	return allBackends
}

func (lb *LoadBalancer) Strategy() LoadBalancingStrategy {
	return lb.strategy
}

func (lb *LoadBalancer) ListenForHealthUpdates() {
	for backendUpdate := range lb.healthUpdateChan {
		lb.UpdateBackend(*backendUpdate)
	}
}

// UpdateBackend updates a backends heath status
func (lb *LoadBalancer) UpdateBackend(update domain.BackendHealthUpdate) {
	if update.IsHealthy {
		lb.addToHealthyBackends(update.Backend, update.GroupId)
	} else {
		lb.removeFromHealthyBackends(update.Backend, update.GroupId)
	}
}

// addToHealthyBackends adds a given backend to the list of healthybackends under a group.
func (lb *LoadBalancer) addToHealthyBackends(backend *domain.Backend, groupId string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	value, ok := lb.healthyBackends.Load(groupId)
	if !ok {
		lb.healthyBackends.Store(groupId, []*domain.Backend{backend})
		return
	}

	hbs, ok := value.([]*domain.Backend)
	if !ok {
		lb.logger.Error("Unexpected type in healthybackends")
		return
	}

	for _, b := range hbs {
		if b.URL == backend.URL {
			return // Backend is already in the healthy list
		}
	}
	// Add to the healthy backends
	lb.healthyBackends.Store(groupId, append(hbs, backend))
}

// removeFromHealthyBackends removes a given backend from the list of healthybackends under a group
func (lb *LoadBalancer) removeFromHealthyBackends(backend *domain.Backend, groupId string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	value, ok := lb.healthyBackends.Load(groupId)
	if !ok {
		return
	}

	hbs, ok := value.([]*domain.Backend)
	if !ok {
		lb.logger.Error("Unexpected type in healthybackends")
		return
	}

	for i, b := range hbs {
		if b.URL == backend.URL {
			lb.healthyBackends.Store(groupId, append(hbs[:i], hbs[i+1:]...))
			return // Backend is already in the healthy list
		}
	}
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

func (lb *LoadBalancer) RouteRequestToGroup(w http.ResponseWriter, r *http.Request, groupId string) {
	startTime := time.Now()
	value, ok := lb.healthyBackends.Load(groupId)
	if !ok {
		lb.logger.Error("No healthy backends found in loadbalancer for group", zap.String("group_id", groupId))
	}
	backends, ok := value.([]*domain.Backend)
	if !ok {
		lb.logger.Error("Unexpected type in healthybackends")
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
