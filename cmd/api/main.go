package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/krispingal/l7lb/internal/domain"
	"github.com/krispingal/l7lb/internal/interfaces/httphandler"
	"github.com/krispingal/l7lb/internal/usecases"
)

func main() {
	backendGroupA := []*domain.Backend{
		{URL: "http://localhost:8081", Alive: true, Health: "/health"},
		{URL: "http://localhost:8083", Alive: true, Health: "/health"},
	}
	backendGroupB := []*domain.Backend{
		{URL: "http://localhost:8082", Alive: true, Health: "/health"},
	}

	apiAlb := usecases.NewLoadBalancer(backendGroupA)
	apiBlb := usecases.NewLoadBalancer(backendGroupB)
	hcA := usecases.NewHealthChecker(backendGroupA, http.DefaultClient)
	hcB := usecases.NewHealthChecker(backendGroupB, http.DefaultClient)

	// Start health checks for backend servers
	go hcA.Start()
	go hcB.Start()

	routes := map[string]*usecases.LoadBalancer{
		"/apiA": apiAlb,
		"/apiB": apiBlb,
	}

	router := httphandler.NewPathRouterWithLB(routes)
	limiter := usecases.NewRateLimiter(100, time.Minute)

	// Load the SSL certificate and key
	certFile := "cert.pem"
	keyFile := "key.pem"

	server := &http.Server{
		Addr:    ":8443",
		Handler: httphandler.NewMiddleware(limiter, router),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	fmt.Println("Load Balancer started at :8443")
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
