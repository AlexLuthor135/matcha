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

func server_run(mux *goji.Mux) {
	err := http.ListenAndServe("0.0.0.0:8000", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/"), helloWorld)
	mux.HandleFunc(pat.Get("/api/health"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "message": "Go backend is running smoothly"}`)
	})

	fmt.Println("Starting Go server on port 8000...")
	server_run(mux)
}