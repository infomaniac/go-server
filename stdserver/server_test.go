package stdserver

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPort(t *testing.T) {
	s := New()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Run(nil, nil, nil)
		assert.NoError(t, err)
	}()

	assert.Equal(t, ":8080", s.getServerAddress())

	time.Sleep(500 * time.Millisecond)
	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", 8080))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	s.Shutdown()
	wg.Wait()
}

func TestPortFromEnv(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	randomPort := rand.Intn((1<<16)-1024) + 1024
	os.Setenv("PORT", fmt.Sprintf("%d", randomPort))

	s := New()
	go func() { err := s.Run(nil, nil, nil); assert.NoError(t, err) }()
	defer s.Shutdown()

	time.Sleep(100 * time.Millisecond)
	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", randomPort))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestInvalidAddress(t *testing.T) {
	s := New()
	s.Address = "invalid address"
	err := s.Run(nil, nil, nil)
	assert.Error(t, err)
	err = s.Shutdown()
	assert.NoError(t, err)
}

func TestServerDefaultHandler(t *testing.T) {
	s := randomPortServer()
	defer s.Shutdown()

	go s.Run(nil, nil, nil)
	time.Sleep(100 * time.Millisecond)

	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "Nothing here.", string(body))
}

func TestServerHandler(t *testing.T) {
	s := randomPortServer()
	defer s.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("index"))
	})

	go func() {
		err := s.Run(mux, nil, nil)
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "index", string(body))
}

func TestServerDefaultHealthzHandler(t *testing.T) {
	s := randomPortServer()
	defer s.Shutdown()

	go func() {
		err := s.Run(nil, nil, nil)
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/healthz", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "up and running.", string(body))
}

func TestServerHealthzHandler(t *testing.T) {
	s := randomPortServer()
	defer s.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("healthz"))
	})

	go func() {
		err := s.Run(nil, mux, nil)
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/healthz", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "healthz", string(body))
}

func TestDebugHandler(t *testing.T) {
	s := randomPortServer()
	s.Debug = true
	defer s.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("debug"))
	})

	go func() {
		err := s.Run(nil, nil, mux)
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	// requests to default hanlder. requests are counted.
	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp, err = http.DefaultClient.Get(fmt.Sprintf("http://%s/asdf", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// debug handler
	resp, err = http.DefaultClient.Get(fmt.Sprintf("http://%s/debug", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, "debug", string(body))

	resp, err = http.DefaultClient.Get(fmt.Sprintf("http://%s/debug/pprof", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.DefaultClient.Get(fmt.Sprintf("http://%s/debug/vars", s.Address))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	data := &struct {
		ServedRequests int64
	}{}
	err = json.Unmarshal(body, data)
	assert.NoError(t, err)

	assert.Equal(t, int64(2), data.ServedRequests)
}

func TestGracefulShutdown(t *testing.T) {
	s := randomPortServer()
	defer s.Shutdown()

	// http.Handler that delays for 30 seconds
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprintf(w, "Hello, world!\n")
	})

	wg := sync.WaitGroup{}
	// start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		runErr := s.Run(mux, nil, nil)
		assert.NoError(t, runErr)
	}()

	// wait for server to start
	time.Sleep(100 * time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, innerErr := http.Get(fmt.Sprintf("http://%s/", s.Address))
		assert.NoError(t, innerErr)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, innerErr := io.ReadAll(resp.Body)
		assert.NoError(t, innerErr)
		assert.Equal(t, "Hello, world!\n", string(body))
	}()

	time.Sleep(500 * time.Millisecond)
	shutdownErr := s.Shutdown()
	assert.NoError(t, shutdownErr)

	wg.Wait()
}

func randomPortServer() *Server {
	rand.Seed(time.Now().UnixNano())
	randomPort := rand.Intn((1<<16)-1024) + 1024
	s := New()
	s.Address = fmt.Sprintf("127.0.0.1:%d", randomPort)
	return s
}
