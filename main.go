package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func main() {
	backendGroupA := []*Backend{
		{URL: "http://localhost:8081", Alive: true, Health: "/health"},
		{URL: "http://localhost:8083", Alive: true, Health: "/health"},
	}
	backendGroupB := []*Backend{
		{URL: "http://localhost:8082", Alive: true, Health: "/health"},
	}

	apiAlb := &LoadBalancer{backends: backendGroupA}
	apiBlb := &LoadBalancer{backends: backendGroupB}

	go apiAlb.healthCheck()
	go apiBlb.healthCheck()

	routes := map[string]*LoadBalancer{
		"/apiA": apiAlb,
		"/apiB": apiBlb,
	}
	router := NewPathRouterWithLB(routes)

	// Load the SSL certificate and key
	certFile := "cert.pem"
	keyFile := "key.pem"

	server := &http.Server{
		Addr:    ":8443",
		Handler: router,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	fmt.Println("Load Balancer started at :8443")
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
