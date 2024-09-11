package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL    string
	Alive  bool
	Health string
}
type LoadBalancer struct {
	backends []*Backend
	current  uint32
}

func (lb *LoadBalancer) getNextBackend() *Backend {
	next := atomic.AddUint32(&lb.current, 1)
	return lb.backends[next%uint32(len(lb.backends))]
}
func (lb *LoadBalancer) healthCheck() {
	for _, backend := range lb.backends {
		go func(b *Backend) {
			for {
				resp, err := http.Get(b.URL + b.Health)
				if err != nil || resp.StatusCode != http.StatusOK {
					b.Alive = false
					log.Printf("Backend %s is down\n", b.URL)
				} else {
					b.Alive = true
					log.Printf("Backend %s is up\n", b.URL)
				}
				time.Sleep(10 * time.Second)
			}
		}(backend)
	}
}

func (lb *LoadBalancer) proxy(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextBackend()

	resp, err := http.Get(backend.URL + r.URL.Path)
	if err != nil {
		http.Error(w, "Backend unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response from backend: %v", err)
		http.Error(w, "Error forwarding response", http.StatusInternalServerError)
	}
}

func main() {
	backends := []*Backend{
		{URL: "http://localhost:8081", Alive: true, Health: "/health"},
		{URL: "http://localhost:8082", Alive: true, Health: "/health"},
	}

	lb := &LoadBalancer{backends: backends}

	go lb.healthCheck()

	// Load the SSL certificate and key
	certFile := "cert.pem"
	keyFile := "key.pem"

	server := &http.Server{
		Addr:    ":8443",
		Handler: http.HandlerFunc(lb.proxy),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	fmt.Println("Load Balancer started at :8443")
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
