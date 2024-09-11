package main

import (
	"io"
	"log"
	"net/http"
	"strings"
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

func NewPathRouterWithLB(routes map[string]*LoadBalancer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, lb := range routes {
			if match := strings.HasPrefix(r.URL.Path, path); match {
				lb.forward(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
}

func (lb *LoadBalancer) forward(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	backend := lb.getNextBackend()

	if !backend.Alive {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		log.Printf("Path: %s | Req method: %s | Backend: %s | Status: %d | Latency: %v\n",
			r.URL.Path, r.Method, backend.URL, http.StatusServiceUnavailable, time.Since(startTime))
		return
	}

	resp, err := http.Get(backend.URL + r.URL.Path)
	if err != nil {
		http.Error(w, "Backend unavailable", http.StatusServiceUnavailable)
		log.Printf("Path: %s | Req method: %s | Backend: %s | Status: %d | Latency: %v\n",
			r.URL.Path, r.Method, backend.URL, http.StatusServiceUnavailable, time.Since(startTime))
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Error forwarding response", http.StatusInternalServerError)
		log.Printf("Path: %s | Req method: %s | Backend: %s | Status: %d | Latency: %v\n",
			r.URL.Path, r.Method, backend.URL, http.StatusInternalServerError, time.Since(startTime))
		return
	}
	log.Printf("Path: %s | Req method: %s | Backend: %s | Status: %d | Latency: %v\n",
		r.URL.Path, r.Method, backend.URL, resp.StatusCode, time.Since(startTime))
}

func (lb *LoadBalancer) getNextBackend() *Backend {
	lb.current = (lb.current + 1) % uint32(len(lb.backends))
	return lb.backends[lb.current]
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
