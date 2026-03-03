package main

import (
	"fmt"
	"log"
	"net/http"

	"goji.io"
	"goji.io/pat"
)

// A simple handler to test that the server is working
func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to your new Go backend!")
}

func main() {
	// 1. Initialize the Goji Router
	mux := goji.NewMux()

	// 2. Define your routes
	mux.HandleFunc(pat.Get("/"), helloWorld)
	mux.HandleFunc(pat.Get("/api/health"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "message": "Go backend is running smoothly"}`)
	})

	// 3. Start the server on port 8000
	fmt.Println("Starting Go server on port 8000...")
	err := http.ListenAndServe("0.0.0.0:8000", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}