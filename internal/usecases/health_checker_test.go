package usecases

import (
	"net/http"
	"net/http/httptest"

	"testing"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
)

func TestHealthChecker_BackendAlive(t *testing.T) {
	// Create a test backend which is down initially
	backend := &domain.Backend{
		URL:    "http://test-backend",
		Health: "/health",
		Alive:  false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Overrride the backend URL to point to the test server
	// with this we simulate backend getting back online
	backend.URL = server.URL

	httpClient := &http.Client{}
	hc := NewHealthChecker([]*domain.Backend{backend}, httpClient)

	go hc.Start()

	// Wait for health check to run
	time.Sleep(1 * time.Second)

	// Verify that backend's Alive status is updated to true
	if !backend.Alive {
		t.Errorf("Expected backend to be alive, but it's not")
	}
}

func TestHealthChecker_BackendDownt(t *testing.T) {
	// Create a test backend which is alive initially
	backend := &domain.Backend{
		URL:    "http://test-backend",
		Health: "/health",
		Alive:  true,
	}

	// Setup an HTTP test server that simulates a failed backend
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Overrride the backend URL to point to the test server
	// with this we simulate backend getting offline
	backend.URL = server.URL

	httpClient := &http.Client{}
	hc := NewHealthChecker([]*domain.Backend{backend}, httpClient)

	go hc.Start()

	// Wait for health check to run
	time.Sleep(1 * time.Second)

	if backend.Alive {
		t.Errorf("Expected backend to be down, but it's alive")
	}
}

func TestHealthChecker_HTTPClientError(t *testing.T) {
	// Create a test backend which is alive initially
	backend := &domain.Backend{
		URL:    "http://test-backend",
		Health: "/health",
		Alive:  true,
	}

	// Overrride the backend URL to point to non-existent URL
	// to simulate an error in the healthcheck
	backend.URL = "http://127.0.0.1:9999"

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))

	defer testServer.Close()

	backend.URL = testServer.URL

	httpClient := &http.Client{}
	hc := NewHealthChecker([]*domain.Backend{backend}, httpClient)

	go hc.Start()

	// Wait for health check to run
	time.Sleep(1 * time.Second)

	if backend.Alive {
		t.Errorf("Expected backend to be down, but it's alive")
	}
}
