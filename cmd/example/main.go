package main

import (
	"log"
	"net/http"
	"time"

	"github.com/infomaniac/go-server"
)

func main() {
	s, err := server.NewServer(handler(), nil)
	if err != nil {
		log.Fatal(err)
	}
	s.ListenAndServe()
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/0", waitHandler(0))
	mux.HandleFunc("/2", waitHandler(2*time.Second))
	mux.HandleFunc("/10", waitHandler(10*time.Second))
	mux.HandleFunc("/30", waitHandler(30*time.Second))
	mux.HandleFunc("/healthz", healthZhandler())
	mux.HandleFunc("/healthZ", healthZhandler())
	return mux
}

func waitHandler(wait time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(wait)
		w.Write([]byte("Hello, world!\n"))
	}
}

func healthZhandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Custom Healthz ok.\n"))
	}
}
