package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	generated "github.com/DistroByte/molecule/internal/generated/go"
)

//go:generate docker run --rm -v $PWD:/spec redocly/cli lint apispec/spec/index.yaml
//go:generate docker run -u 1000:1000 --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/apispec/spec/index.yaml -g go-server -o /local/internal/generated -c /local/apispec/server-config.yaml

var (
	nomadUrl = "http://zeus.internal:4646"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var nomadService v1.NomadServiceInterface

	if os.Getenv("PROD") == "true" {
		nomadClient, err := api.NewClient(&api.Config{Address: nomadUrl})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Nomad client")
			return
		}
		nomadService = v1.NewNomadService(nomadClient)
	} else {
		nomadService = v1.NewMockNomadService()
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("running locally in debug mode")
	}

	moleculeAPIService := v1.NewMoleculeAPIService(nomadService)
	moleculeAPIController := generated.NewDefaultAPIController(moleculeAPIService)

	// Add X-API-KEY authentication
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("API_KEY environment variable is required")
	}

	r := chi.NewRouter()
	r.Use(requestIDMiddleware)
	r.Use(zerologMiddleware)

	// Serve static HTML content from the root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	r.Route("/static", func(staticRouter chi.Router) {
		staticRouter.Use(middleware.NoCache)
		staticRouter.Handle("/*", http.StripPrefix("/static", http.FileServer(http.Dir("./web"))))
	})

	r.Get("/api/spec.json", func(w http.ResponseWriter, r *http.Request) {
		specPath := "./apispec/spec/index.yaml"
		yamlData, err := os.ReadFile(specPath)
		if err != nil {
			http.Error(w, "Failed to read OpenAPI spec", http.StatusInternalServerError)
			return
		}

		var jsonData map[string]interface{}
		if err := yaml.Unmarshal(yamlData, &jsonData); err != nil {
			http.Error(w, "Failed to parse OpenAPI spec", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(jsonData); err != nil {
			http.Error(w, "Failed to encode OpenAPI spec as JSON", http.StatusInternalServerError)
		}
	})

	apiRouter := chi.NewRouter()
	apiRouter.Use(apiKeyAuthHandler(apiKey))
	for _, route := range moleculeAPIController.Routes() {
		if requiresAuth(route.Pattern) {
			apiRouter.Method(route.Method, route.Pattern, route.HandlerFunc)
		} else {
			r.Method(route.Method, route.Pattern, route.HandlerFunc)
		}
	}
	r.Mount("/", apiRouter)

	log.Info().Msgf("Starting server on :8080")
	log.Fatal().Err(http.ListenAndServe(":8080", r))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		logger := log.With().Str("request_id", requestID).Logger()
		r = r.WithContext(logger.WithContext(r.Context()))
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func zerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		logger := zerolog.Ctx(r.Context())
		defer func() {
			logger.Info().
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

func apiKeyAuthHandler(key string) func(http.Handler) http.Handler {
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

func requiresAuth(pattern string) bool {
	log.Debug().Msgf("Checking if route %s requires authentication", pattern)
	authenticatedRoutes := []string{
		"/v1/services/{service}/alloc-restart",
	}

	for _, route := range authenticatedRoutes {
		if route == pattern {
			log.Debug().Msgf("Route %s requires authentication", pattern)
			return true
		}
	}
	return false
}
