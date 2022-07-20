package server

import (
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	os.Setenv("PORT", "1234")

	h := func(w http.ResponseWriter, r *http.Request) {}
	http.HandleFunc("/", h)
	s, err := NewServer(http.DefaultServeMux, nil)
	assert.NoError(t, err)

	assert.Equal(t, ":1234", s.Address)
	assert.NotNil(t, s.Handler)
	assert.NotNil(t, s.HealthZHandler)
}

func TestNewServerNilHandler(t *testing.T) {
	s, err := NewServer(nil, nil)
	assert.ErrorIs(t, err, ErrNoHandler)
	assert.Nil(t, s)
}

func TestNewServerDefaultPort(t *testing.T) {
	os.Unsetenv("PORT")

	s, err := NewServer(http.DefaultServeMux, nil)
	assert.NoError(t, err)

	assert.Equal(t, ":8080", s.Address)
}

func TestListenAndServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("index")) })
	s, _ := NewServer(mux, nil)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.ListenAndServe()
		assert.ErrorIs(t, err, http.ErrServerClosed)
	}()

	time.Sleep(time.Second)
	err := s.Shutdown()
	assert.NoError(t, err, "shutdown")
	wg.Wait()
}

func TestStandardHandler(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ABC")) }
	health := defaultHealthZ()
	mux := http.NewServeMux()
	mux.HandleFunc("/", h)

	s, err := NewServer(mux, health)
	assert.NoError(t, err)

	assert.HTTPBodyContains(t, s.standardHandler().ServeHTTP, "GET", "/healthZ", nil, "OK")
	assert.HTTPBodyContains(t, s.standardHandler().ServeHTTP, "GET", "/", nil, "ABC")
}

func TestDefaultHealthZ(t *testing.T) {
	assert.HTTPSuccess(t, defaultHealthZ().ServeHTTP, "GET", "", nil)
}
