package main

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from backend 2")
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", apiHandler)
	fmt.Println("Backend 1 listening on :8082")
	http.ListenAndServe(":8082", nil)
}
