package stdserver

import (
	"context"
	"expvar"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	// ctx     context.Context
	Address string
	server  http.Server
	stats   ServerStats
	Debug   bool
}

type ServerStats struct {
	servedRequests  expvar.Int
	openConnections expvar.Int
}

func New() *Server {
	return &Server{}
}

// Server.Run() starts the server
func (s *Server) Run(handler, healthzHandler, debugHandler http.Handler) error {
	if handler == nil {
		handler = &DefaultHandler{}
	}
	if healthzHandler == nil {
		healthzHandler = &HealthzHandler{}
	}

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthzHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.stats.servedRequests.Add(1)
		handler.ServeHTTP(w, r)
	})

	if s.Debug {
		expvar.Publish("openConnections", &s.stats.openConnections)
		expvar.Publish("servedRequests", &s.stats.servedRequests)
		if debugHandler != nil {
			mux.Handle("/debug/", debugHandler)
		}
		mux.Handle("/debug/vars/", expvar.Handler())
		mux.HandleFunc("/debug/pprof/", pprof.Index)
	}

	s.server.Addr = s.getServerAddress()
	s.server.Handler = mux
	s.server.IdleTimeout = 60 * time.Second
	log.Printf("Starting server on %v", s.server.Addr)

	// track open connections
	s.server.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			s.stats.openConnections.Add(1)
		case http.StateClosed:
			s.stats.openConnections.Add(-1)
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.ShutdownOnInterrupt(wg)

	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	wg.Wait()
	return nil
}

// Server.getServerAddress returns the port from the server value, environment or a default value.
func (s *Server) getServerAddress() string {
	if s.Address != "" {
		return s.Address
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer s.server.Close()
	return s.server.Shutdown(ctx)
}

func (s *Server) ShutdownOnInterrupt(wg *sync.WaitGroup) {
	defer wg.Done()
	sigInt := make(chan os.Signal, 1)
	signal.Notify(sigInt, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigInt
	log.Printf("Signal %v received, shutting down server", sig)
	err := s.Shutdown()
	if err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}
