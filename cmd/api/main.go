package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/krispingal/l7lb/internal/infrastructure"
	"github.com/krispingal/l7lb/internal/interfaces/httphandler"
	"github.com/krispingal/l7lb/internal/usecases"
	"github.com/krispingal/l7lb/internal/usecases/loadbalancing"
)

func main() {
	config, err := infrastructure.LoadConfig("config")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	loadBalancers := loadbalancing.CreateLoadBalanacers(config)

	// Start health checks for backend group
	for _, lb := range loadBalancers {
		backends := lb.Backends()
		hc := usecases.NewHealthChecker(backends, http.DefaultClient)
		go hc.Start()
	}

	router := httphandler.NewPathRouterWithLB(loadBalancers)
	fixedWindowLimiter := usecases.NewFixedWindowRateLimiter(100, time.Minute)

	// Load the SSL certificate and key
	certFile := "cert.pem"
	keyFile := "key.pem"

	server := &http.Server{
		Addr:    config.LoadBalancer.Address,
		Handler: httphandler.NewMiddleware(fixedWindowLimiter, router),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	log.Printf("Load Balancer started at %s\n", config.LoadBalancer.Address)
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
