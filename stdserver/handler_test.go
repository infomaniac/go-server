package stdserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHandler(t *testing.T) {
	h := &DefaultHandler{}
	assert.HTTPStatusCode(t, h.ServeHTTP, "GET", "/", nil, 404)
	assert.HTTPBodyContains(t, h.ServeHTTP, "GET", "/", nil, "Nothing here.")
}

func TestHealthzHandler(t *testing.T) {
	h := &HealthzHandler{}
	assert.HTTPSuccess(t, h.ServeHTTP, "GET", "/", nil)
	assert.HTTPBodyContains(t, h.ServeHTTP, "GET", "/", nil, "up and running.")
}
