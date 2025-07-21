package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	server := New("localhost", 8080)

	assert.NotNil(t, server)
	assert.Equal(t, "localhost", server.host)
	assert.Equal(t, 8080, server.port)
	assert.NotNil(t, server.Router())
}

func TestNew_DefaultPort(t *testing.T) {
	server := New("localhost", 0)

	assert.NotNil(t, server)
	assert.Equal(t, "localhost", server.host)
	assert.Equal(t, 8080, server.port) // Should default to 8080
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create a test handler that checks for request ID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that X-Request-ID header is set
		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("Request ID header not set")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap it with RequestID middleware
	handler := RequestIDMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}

func TestAPIKeyAuthMiddleware_ValidKey(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	middleware := APIKeyAuthMiddleware("valid-api-key")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-KEY", "valid-api-key")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "success", recorder.Body.String())
}

func TestAPIKeyAuthMiddleware_InvalidKey(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	middleware := APIKeyAuthMiddleware("valid-api-key")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-KEY", "invalid-key")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Unauthorized")
}

func TestAPIKeyAuthMiddleware_MissingKey(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	middleware := APIKeyAuthMiddleware("valid-api-key")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	// Don't set X-API-KEY header
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Unauthorized")
}

func TestRequiresAuth(t *testing.T) {
	testCases := []struct {
		pattern  string
		expected bool
	}{
		{"/v1/services/{service}/alloc-restart", true},
		{"/v1/urls", false},
		{"/v1/services", false},
		{"/health", false},
		{"/", false},
	}

	for _, tc := range testCases {
		t.Run(tc.pattern, func(t *testing.T) {
			result := RequiresAuth(tc.pattern)
			assert.Equal(t, tc.expected, result)
		})
	}
}
