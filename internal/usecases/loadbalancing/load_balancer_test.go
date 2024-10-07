package loadbalancing

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krispingal/l7lb/internal/domain"
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

func TestLoadBalancerRouteRequest(t *testing.T) {
	// Create a mock backend using httptest.Server to simulate the backend server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Simulate a successful response
	}))
	defer mockServer.Close()
	backend := &domain.Backend{
		URL:   mockServer.URL,
		Alive: true,
	}

	mockStrategy := &MockStrategy{backend: backend}
	lb := NewLoadBalancer([]*domain.Backend{backend}, mockStrategy)

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
}

func TestLoadBalancerRouteRequestUnavailableBackend(t *testing.T) {
	backend := &domain.Backend{
		URL:   "http://mock-backend",
		Alive: false,
	}

	mockStrategy := &MockStrategy{backend: backend}
	lb := NewLoadBalancer([]*domain.Backend{backend}, mockStrategy)

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Result().StatusCode)
	}
}

func TestLoadBalancerRouteRequestStrategyError(t *testing.T) {
	backend := &domain.Backend{
		URL:   "http://mock-backend",
		Alive: true,
	}
	mockStrategy := &MockStrategy{err: errors.New("strategy error")}
	lb := NewLoadBalancer([]*domain.Backend{backend}, mockStrategy)

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()

	lb.RouteRequest(w, req)

	if w.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Result().StatusCode)
	}
}