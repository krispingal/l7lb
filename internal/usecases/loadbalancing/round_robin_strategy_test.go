package loadbalancing

import (
	"testing"

	"github.com/krispingal/l7lb/internal/domain"
)

func TestRoundRobinStrategy(t *testing.T) {
	backends := []*domain.Backend{
		{URL: "http://localhost:8081", Health: "/health"},
		{URL: "http://localhost:8082", Health: "/health"},
		{URL: "http://localhost:8083", Health: "/health"},
	}
	strategy := NewRoundRobinStrategy()
	selected, err := strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8081" {
		t.Errorf("Expected first backend got %s", selected.URL)
	}
	selected, err = strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8082" {
		t.Errorf("Expected second backend got %s", selected.URL)
	}
	selected, err = strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8083" {
		t.Errorf("Expected third backend got %s", selected.URL)
	}
	selected, err = strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8081" {
		t.Errorf("Expected first backend got %s", selected.URL)
	}
}
