package main

import (
	"log"
	"net/http"
	"time"

	"github.com/infomaniac/go-server/stdserver"
)

func main() {
	log.Default().SetFlags(log.LstdFlags | log.Lmicroseconds)

	s := stdserver.New()
	s.Debug = true

	hello := HelloHandler()
	everythingOk := &EverythingOkHandler{}

	err := s.Run(hello, everythingOk, nil)

	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func HelloHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello/world", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!\n"))
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Unknow!\n"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Stranger!\n"))
	})
	mux.HandleFunc("/wait20", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Second)
		w.Write([]byte("ok!\n"))
	})
	mux.HandleFunc("/wait60", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Second)
		w.Write([]byte("ok!\n"))
	})
	return mux
}

// EverythingOkHandler is a simple HttpHandler that responds with "Everything Ok"
type EverythingOkHandler struct{}

func (h *EverythingOkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Everything Ok"))
}
