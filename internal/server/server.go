package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/DistroByte/molecule/logger"
)

// Server represents the HTTP server
type Server struct {
	router chi.Router
	host   string
	port   int
}

// New creates a new server instance
func New(host string, port int) *Server {
	if port == 0 {
		port = 8080
	}

	r := chi.NewRouter()
	r.Use(RequestIDMiddleware)
	r.Use(ZerologMiddleware)

	return &Server{
		router: r,
		host:   host,
		port:   port,
	}
}

// Router returns the chi router for route configuration
func (s *Server) Router() chi.Router {
	return s.router
}

// Start starts the HTTP server
func (s *Server) Start() error {
	serverHost := s.host
	if serverHost == "" {
		serverHost = "localhost"
	}

	logger.Log.Info().Msgf("starting server on http://%s:%d", serverHost, s.port)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.host, s.port), s.router)
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		httpLogger := logger.Log.With().Str("request_id", requestID).Logger()
		r = r.WithContext(httpLogger.WithContext(r.Context()))
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

// ZerologMiddleware logs HTTP requests
func ZerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		httpLogger := zerolog.Ctx(r.Context())
		defer func() {
			httpLogger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Str("remote", r.RemoteAddr).
				Dur("duration", time.Since(start)).
				Msg("handled request")
		}()

		next.ServeHTTP(ww, r)
	})
}

// APIKeyAuthMiddleware creates middleware for API key authentication
func APIKeyAuthMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-KEY") != key {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequiresAuth determines if a route pattern requires authentication
func RequiresAuth(pattern string) bool {
	logger.Log.Debug().Msgf("checking if route %s requires authentication", pattern)
	authenticatedRoutes := []string{
		"/v1/services/{service}/alloc-restart",
	}

	for _, route := range authenticatedRoutes {
		if route == pattern {
			logger.Log.Debug().Msgf("route %s requires authentication", pattern)
			return true
		}
	}
	return false
}