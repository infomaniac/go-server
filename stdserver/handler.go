package stdserver

import "net/http"

type DefaultHandler struct{}

// DefaultHandler.ServeHTTP is the default handler.
func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Nothing here."))
}

type HealthzHandler struct {
}

// HealthzHandler.ServeHTTP is the default healthz handler.
func (h *HealthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("up and running."))
}
