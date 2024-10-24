package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	response := os.Getenv("BACKEND_RESPONSE")
	if response == "" {
		response = "Hello from backend!"
	}
	fmt.Fprintf(w, response)
}

func main() {
	port := flag.String("port", "8080", "Port to run backend server on")
	flag.Parse()

	h2Server := &http2.Server{}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", apiHandler)

	handler := h2c.NewHandler(mux, h2Server)

	fmt.Printf("Backend server listening on :%s\n", *port)
	if err := http.ListenAndServe(":"+*port, handler); err != nil {
		fmt.Printf("Error starting server on port %s: %v\n", *port, err)
	}

}
