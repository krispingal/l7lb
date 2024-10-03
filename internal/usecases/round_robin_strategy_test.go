package usecases

import (
	"testing"

	"github.com/krispingal/l7lb/internal/domain"
)

func TestRoundRobinStrategy(t *testing.T) {
	backends := []*domain.Backend{
		{URL: "http://localhost:8081", Alive: true},
		{URL: "http://localhost:8082", Alive: true},
		{URL: "http://localhost:8083", Alive: true},
	}
	strategy := NewRoundRobinStrategy()
	lb := NewLoadBalancer(backends, strategy)
	selected, err := lb.strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8081" {
		t.Errorf("Expected first backend got %s", selected.URL)
	}
	selected, err = lb.strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8082" {
		t.Errorf("Expected second backend got %s", selected.URL)
	}
	selected, err = lb.strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8083" {
		t.Errorf("Expected third backend got %s", selected.URL)
	}
	selected, err = lb.strategy.GetNextBackend(backends)
	if err != nil {
		t.Errorf("Did not expect any errors")
	} else if selected.URL != "http://localhost:8081" {
		t.Errorf("Expected first backend got %s", selected.URL)
	}
}
