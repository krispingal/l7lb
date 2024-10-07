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
	"github.com/krispingal/l7lb/internal/usecases/ratelimiting"
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
	var rateLimiter ratelimiting.RateLimiterInterface
	switch config.RateLimiter.Type {
	case "none":
		rateLimiter = ratelimiting.NoOpRateLimiter{}
	case "fixed_window":
		windowDuration, err := time.ParseDuration(config.RateLimiter.Window)
		if err != nil {
			log.Fatalf("Invalid fixed window ratelimiter window duration: %v", err)
		}
		// fixed window rate limiter
		rateLimiter = ratelimiting.NewFixedWindowRateLimiter(config.RateLimiter.Limit, windowDuration)
	default:
		log.Fatalf("Invalid rate limiter type: %s", config.RateLimiter.Type)
	}
	// Load the SSL certificate and key
	certFile := "cert.pem"
	keyFile := "key.pem"

	// TLSConfig with optimized settings for security and performance
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Allows TLS 1.2 fallback
		MaxVersion: tls.VersionTLS13, // Prefer TLS 1.3 for security and performance

		// Specify secure and performant cipher suites for TLS 1.2 (if needed)
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	server := &http.Server{
		Addr:      config.LoadBalancer.Address,
		Handler:   httphandler.NewMiddleware(rateLimiter, router),
		TLSConfig: tlsConfig,
	}

	log.Printf("Load Balancer started at %s\n", config.LoadBalancer.Address)
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
