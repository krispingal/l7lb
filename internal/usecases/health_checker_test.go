package usecases

import (
	"net/http"
	"net/http/httptest"

	"testing"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"go.uber.org/zap/zaptest"
)

type MockBackendRegistry struct {
	// updateHealthInvoked tracks whether UpdateHealth was called.
	updateHealthInvoked bool
	updatedStatus       domain.BackendStatus
}

func (m *MockBackendRegistry) UpdateHealth(status domain.BackendStatus) error {
	m.updateHealthInvoked = true
	m.updatedStatus = status // Capture the status passed to UpdateHealth
	return nil
}

func (m *MockBackendRegistry) Subscribe(backendId uint64) <-chan domain.BackendStatus {
	return nil
}

func (m *MockBackendRegistry) GetBackendById(backendId uint64) (domain.Backend, bool) {
	return domain.Backend{}, false
}

func (m *MockBackendRegistry) AddBackendToRegistry(backend domain.Backend) {}

func setupTest(t *testing.T, healthyBackend bool, markAsHealthy bool) (*MockBackendRegistry, *HealthChecker, *httptest.Server, *domain.Backend) {
	testBackend := &domain.Backend{
		Id:     11,
		URL:    "http://test-backend",
		Health: "/health",
	}
	testLogger := zaptest.NewLogger(t)

	// Set up an HTTP test server.
	var server *httptest.Server
	if healthyBackend {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	} else {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
	}

	// Point the test backend URL to the test server.
	testBackend.URL = server.URL

	// Create a mock registry.
	mockRegistry := &MockBackendRegistry{}

	// Create an HTTP client.
	httpClient := &http.Client{}

	// Create a health checker with a 100ms interval and 1s timeout.
	hc := NewHealthChecker(100*time.Millisecond, 1*time.Second, mockRegistry, httpClient, testLogger)

	if markAsHealthy {
		hc.healthySet.Store(server.URL, testBackend)
	}

	return mockRegistry, hc, server, testBackend
}

func TestHealthChecker_BackendBecomesAlive(t *testing.T) {
	mockRegistry, hc, server, testBackend := setupTest(t, true, false)
	defer server.Close()

	go hc.Start()
	hc.AddBackend(testBackend)

	// Wait for 2 health check cycles to ensure the health check has run
	time.Sleep(250 * time.Millisecond)

	// Verify that UpdateHealth was called
	if !mockRegistry.updateHealthInvoked {
		t.Errorf("Expected UpdateHealth to be called, but it wasn't")
	}

	// Verify that the correct backend status was passed to UpdateHealth
	expectedStatus := domain.BackendStatus{Id: 11, IsHealthy: true}
	if mockRegistry.updatedStatus != expectedStatus {
		t.Errorf("Expected UpdateHealth to be called with %+v, but got %+v", expectedStatus, mockRegistry.updatedStatus)
	}
}

func TestHealthChecker_BackendAliveInitially(t *testing.T) {
	mockRegistry, hc, server, testBackend := setupTest(t, true, true)
	defer server.Close()
	// Start the health checker.
	go hc.Start()

	// Add the test backend.
	hc.AddBackend(testBackend)

	// Wait for 2 health check cycles.
	time.Sleep(250 * time.Millisecond)

	// Verify UpdateHealth was not called.
	if mockRegistry.updateHealthInvoked {
		t.Errorf("Expected UpdateHealth to not be called, but it was")
	}
}

func TestHealthChecker_BackendDownInitially(t *testing.T) {
	mockRegistry, hc, server, testBackend := setupTest(t, false, false)
	defer server.Close()

	go hc.Start()
	hc.serverChan <- testBackend

	// Wait for 2 health check cycles to ensure the health check has run
	time.Sleep(250 * time.Millisecond)

	// Verify that UpdateHealth was not called
	if mockRegistry.updateHealthInvoked {
		t.Errorf("Expected UpdateHealth to not be called, but it was")
	}
}

func TestHealthChecker_BackendBecomesUnhealthy(t *testing.T) {
	mockRegistry, hc, server, testBackend := setupTest(t, false, true)
	defer server.Close()

	go hc.Start()
	hc.serverChan <- testBackend

	// Wait for 2 health check cycles.
	time.Sleep(250 * time.Millisecond)

	// Verify UpdateHealth was called.
	if !mockRegistry.updateHealthInvoked {
		t.Errorf("Expected UpdateHealth to be called, but it was not")
	}

	// Verify the correct backend status was passed to UpdateHealth.
	expectedStatus := domain.BackendStatus{Id: 11, IsHealthy: false}
	if mockRegistry.updatedStatus != expectedStatus {
		t.Errorf("Expected UpdateHealth to be called with %+v, but got %+v", expectedStatus, mockRegistry.updatedStatus)
	}
}

func TestHealthChecker_AddBackend(t *testing.T) {
	// Create a health checker
	mockRegistry := &MockBackendRegistry{}
	httpClient := &http.Client{}
	testLogger := zaptest.NewLogger(t)
	testChan := make(chan *domain.Backend)

	hc := NewHealthChecker(100*time.Millisecond, 1*time.Second, mockRegistry, httpClient, testLogger)
	hc.serverChan = testChan

	// Create a test backend
	testBackend := &domain.Backend{
		URL: "http://test-backend",
	}

	done := make(chan struct{})
	go func() {
		backend := <-testChan
		if backend.URL != testBackend.URL {
			t.Errorf("Expected backend URL to be %s, but got %s", testBackend.URL, backend.URL)
		}
		close(done)
	}()
	// Add the test backend
	hc.AddBackend(testBackend)

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout waiting for backend on server channel")
	}
}
