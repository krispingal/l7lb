package loadbalancing

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type MockStrategy struct {
	backend *domain.Backend
	err     error
}

func (ms *MockStrategy) GetNextBackend([]*domain.Backend) (*domain.Backend, error) {
	if ms.err != nil {
		return nil, ms.err
	}
	return ms.backend, nil
}

var retryCount int

func TestLoadBalancerRouteRequestWithRetries(t *testing.T) {
	// Create a mock backend that fails the first two times & succeeds the third time
	retryCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 3 {
			w.WriteHeader(http.StatusInternalServerError) // Simulate a backend error
		} else {
			w.WriteHeader(http.StatusOK) // Simulate a successful response
		}
	}))
	defer mockServer.Close()
	testLogger := zaptest.NewLogger(t)
	backend := &domain.Backend{
		URL:   mockServer.URL,
		Alive: true,
	}

	mockStrategy := &MockStrategy{backend: backend}
	// Manually inject healthy backends
	lb := &LoadBalancer{
		strategy:        mockStrategy,
		logger:          testLogger,
		healthyBackends: []*domain.Backend{backend},
	}

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
}

func TestLoadBalancerRouteRequestUnavailableBackendWithRetries(t *testing.T) {
	retryCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		w.WriteHeader(http.StatusServiceUnavailable) // Simulate a failing backend
	}))
	defer mockServer.Close()
	testLogger := zaptest.NewLogger(t)

	backend := &domain.Backend{
		URL:   mockServer.URL,
		Alive: true,
	}

	mockStrategy := &MockStrategy{backend: backend}
	lb := &LoadBalancer{
		strategy:        mockStrategy,
		logger:          testLogger,
		healthyBackends: []*domain.Backend{backend},
	}

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	// Check if the final response after retries is a 5xx failure
	if w.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Result().StatusCode)
	}

	if retryCount != 3 {
		t.Errorf("expected 3 retries, got %d", retryCount)
	}
}

func TestLoadBalancerRouteRequestStrategyError(t *testing.T) {
	backend := &domain.Backend{
		URL:   "http://mock-backend",
		Alive: true,
	}
	testLogger := zaptest.NewLogger(t)
	mockStrategy := &MockStrategy{err: errors.New("strategy error")}
	lb := &LoadBalancer{
		strategy:        mockStrategy,
		logger:          testLogger,
		healthyBackends: []*domain.Backend{backend},
	}

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Result().StatusCode)
	}
}

func BenchmarkRouteRequest(b *testing.B) {
	// Register mock backends
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Simulate a successful response
	}))
	defer mockServer.Close()

	testLogger := zap.NewNop()
	backend := &domain.Backend{
		URL:   mockServer.URL,
		Alive: true,
	}
	mockStrategy := &MockStrategy{backend: backend}

	// Set up the LoadBalancer with a healthy backend
	lb := &LoadBalancer{
		strategy:        mockStrategy,
		logger:          testLogger,
		healthyBackends: []*domain.Backend{backend},
	}
	// Mock an HTTP request
	// Create a dummy request
	req := httptest.NewRequest("GET", "http://localhost/api", nil)

	for i := 0; i < b.N; i++ {
		// Benchmark the RouteRequest method
		recorder := httptest.NewRecorder()
		start := time.Now()

		lb.RouteRequest(recorder, req)

		// Measure elapsed time
		duration := time.Since(start)
		b.Logf("Request completed in %v with status %d", duration, recorder.Result().StatusCode)
	}
}
