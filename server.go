package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	ctx            context.Context
	Address        string
	Handler        http.Handler
	HealthZHandler http.Handler

	server *http.Server
}

var (
	ErrNoHandler = errors.New("no handler")
)

func NewServer(handler, healthZhandler http.Handler) (*Server, error) {
	if handler == nil {
		return nil, ErrNoHandler
	}
	if healthZhandler == nil {
		healthZhandler = defaultHealthZ()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port

	s := &Server{
		ctx:            context.Background(),
		Address:        address,
		Handler:        handler,
		HealthZHandler: healthZhandler,
	}

	return s, nil
}

func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:    s.Address,
		Handler: s.standardHandler(),
	}
	s.server = srv

	idleConnsClosed := make(chan struct{})
	registerSignals(srv, idleConnsClosed)
	err := srv.ListenAndServe()
	select {
	case <-idleConnsClosed:
	case <-time.After(1 * time.Second):
		log.Printf("Timeout shutting down server. Forcing exit.")
	}
	return err
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func (s *Server) standardHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthZ", s.HealthZHandler)
	mux.Handle("/", s.Handler)
	return mux
}

func registerSignals(s *http.Server, idleConnsClosed chan struct{}) {
	sigInt := make(chan os.Signal, 1)
	signal.Notify(sigInt, syscall.SIGINT, syscall.SIGTERM)
	go func(s *http.Server) {
		sig := <-sigInt
		log.Printf("Signal %v received, shutting down server", sig)
		err := s.Shutdown(context.Background())
		if err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
		close(idleConnsClosed)
	}(s)
}

func defaultHealthZ() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}
