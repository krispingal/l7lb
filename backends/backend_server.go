package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
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
	fmt.Printf("Received req with method: %s and query : %s\n", r.Method, r.URL.RawQuery)
}

func main() {
	port := flag.String("port", "8080", "Port to run backend server on")
	flag.Parse()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", apiHandler)

	fmt.Printf("Backend server listening on :%s\n", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fmt.Printf("Error starting server on port %s: %v\n", *port, err)
	}

}
