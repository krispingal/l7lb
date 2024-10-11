package loadbalancing

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krispingal/l7lb/internal/domain"
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
	retryCount = 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 3 {
			w.WriteHeader(http.StatusInternalServerError) // Simulate a backend error
		} else {
			w.WriteHeader(http.StatusOK) // Simulate a successful response
		}
	}))
	testLogger := zaptest.NewLogger(t)
	defer mockServer.Close()
	backend := &domain.Backend{
		URL:   mockServer.URL,
		Alive: true,
	}

	mockStrategy := &MockStrategy{backend: backend}
	lb := NewLoadBalancer([]*domain.Backend{backend}, mockStrategy, testLogger)

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
}

func TestLoadBalancerRouteRequestUnavailableBackendWithRetries(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	retryCount = 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		w.WriteHeader(http.StatusServiceUnavailable) // Simulate a failing backend
	}))
	defer mockServer.Close()

	backend := &domain.Backend{
		URL:   mockServer.URL,
		Alive: true,
	}

	mockStrategy := &MockStrategy{backend: backend}
	lb := NewLoadBalancer([]*domain.Backend{backend}, mockStrategy, testLogger)

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
	lb := NewLoadBalancer([]*domain.Backend{backend}, mockStrategy, testLogger)

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Result().StatusCode)
	}
}
