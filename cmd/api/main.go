package main

import (
	"crypto/rand"
	"crypto/tls"
	"net/http"
	"time"

	// _ "net/http/pprof"

	"github.com/krispingal/l7lb/internal/infrastructure"
	"github.com/krispingal/l7lb/internal/interfaces/httphandler"
	"github.com/krispingal/l7lb/internal/usecases"
	"github.com/krispingal/l7lb/internal/usecases/loadbalancing"
	"github.com/krispingal/l7lb/internal/usecases/ratelimiting"
)

func main() {
	infrastructure.InitLogger()
	logger := infrastructure.Logger
	sugar := infrastructure.Logger.Sugar()

	// go func() {
	// 	sugar.Info(http.ListenAndServe("localhost:6060", nil))
	// }()
	config, err := infrastructure.LoadConfig("config")
	if err != nil {
		sugar.Fatalf("Error loading config: %v", err)
	}
	transport := &http.Transport{
		MaxIdleConns:        50, // Maximum number of idle connections
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false, // Ensure keep-alives are enabled for connection reuse
	}
	pooledClient := &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}
	hc_healthy_freq, err1 := time.ParseDuration(config.HealthChecker.HealthyServerFrequency)
	hc_unhealthy_freq, err2 := time.ParseDuration(config.HealthChecker.UnhealthyServerFrequency)
	if err1 != nil || err2 != nil {
		sugar.Fatalf("Invalid time duration provided for healthchecker frequency: %v, %v", err1, err2)
	}

	hc := usecases.NewHealthChecker(hc_healthy_freq, hc_unhealthy_freq, pooledClient, logger)

	loadBalancers := loadbalancing.CreateLoadBalancers(config, logger)

	// Start health checks for backend group
	for _, lb := range loadBalancers {
		lb.SubscribeToHealthChecker(hc) // Subscribe each LB to the healthchecker's channel
		for _, backend := range lb.Backends() {
			hc.AddBackend(backend)
		}
	}
	hc.Start()

	router := httphandler.NewPathRouterExactPathWithLB(loadBalancers)
	var rateLimiter ratelimiting.RateLimiterInterface
	switch config.RateLimiter.Type {
	case "none":
		rateLimiter = ratelimiting.NoOpRateLimiter{}
	case "fixed_window":
		windowDuration, err := time.ParseDuration(config.RateLimiter.Window)
		if err != nil {
			sugar.Fatalf("Invalid fixed window ratelimiter window duration: %v", err)
		}
		// fixed window rate limiter
		rateLimiter = ratelimiting.NewFixedWindowRateLimiter(config.RateLimiter.Limit, windowDuration)
	default:
		sugar.Fatalf("Invalid rate limiter type: %s", config.RateLimiter.Type)
	}

	// Generate a session ticket key for session resumption
	sessionTicketKey := [32]byte{}
	if _, err := rand.Read(sessionTicketKey[:]); err != nil {
		sugar.Fatalf("failed to generate a session ticket key: %v", err)
	}

	// TLSConfig with optimized settings for security and performance
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,           // Allows TLS 1.2 fallback
		MaxVersion: tls.VersionTLS13,           // Prefer TLS 1.3 for security and performance
		NextProtos: []string{"h2", "http/1.1"}, // Enable HTTP/2

		SessionTicketsDisabled: false,
		SessionTicketKey:       sessionTicketKey,

		// Specify secure and performant cipher suites for TLS 1.2 (if needed)
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	server := &http.Server{
		Addr:      config.LoadBalancer.Address,
		Handler:   httphandler.NewMiddleware(rateLimiter, router, logger),
		TLSConfig: tlsConfig,
	}

	sugar.Infof("Load Balancer started at %s", config.LoadBalancer.Address)
	sugar.Fatal(server.ListenAndServeTLS(config.LoadBalancer.CertFile, config.LoadBalancer.KeyFile))
}
