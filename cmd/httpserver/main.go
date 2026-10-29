package main

import (
	"fmt"
	"httpServer/internal/http1"
	"httpServer/internal/mux1"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const addr = ":8080"

func main() {
	mux := mux1.NewMyMux()

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the homepage!")
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Status: Systems Operational")
	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "deal with post")
	})

	err := http1.ListenAndServe(addr, mux)
	if err != nil {
		panic("ijijij")
	}
}
