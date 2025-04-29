package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	generated "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/DistroByte/molecule/logger"
)

//go:generate docker run --rm -v $PWD:/spec redocly/cli lint apispec/spec/index.yaml
//go:generate docker run -u 1000:1000 --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/apispec/spec/index.yaml -g go-server -o /local/internal/generated -c /local/apispec/server-config.yaml

type Config struct {
	Nomad struct {
		Address string `yaml:"address"`
	} `yaml:"nomad"`

	StandardURLs []struct {
		Service string `yaml:"service"`
		URL     string `yaml:"url"`
	} `yaml:"standard_urls"`
}

func main() {

	switch os.Getenv("LEVEL") {
	case "trace":
		logger.InitLogger(zerolog.TraceLevel)
	case "debug":
		logger.InitLogger(zerolog.DebugLevel)
	case "info":
		logger.InitLogger(zerolog.InfoLevel)
	case "warn":
		logger.InitLogger(zerolog.WarnLevel)
	case "error":
		logger.InitLogger(zerolog.ErrorLevel)
	case "fatal":
		logger.InitLogger(zerolog.FatalLevel)
	case "panic":
		logger.InitLogger(zerolog.PanicLevel)
	default:
		logger.InitLogger(zerolog.InfoLevel)
	}

	logger.Log.Info().Msg("Logger initialized")

	var nomadService v1.NomadServiceInterface

	if os.Getenv("PROD") == "true" {
		configFilePath := "./config.yaml"
		config, err := loadConfig(configFilePath)

		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to load config file")
			return
		}

		logger.Log.Debug().Any("urls", config.StandardURLs).Msg("Loaded standard URLs from YAML")
		if len(config.StandardURLs) == 0 {
			logger.Log.Warn().Msg("No standard URLs found in YAML file")
		}

		nomadClient, err := api.NewClient(&api.Config{Address: config.Nomad.Address})
		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to create api client")
			return
		}

		standardURLsMap := make(map[string]string)
		for _, entry := range config.StandardURLs {
			standardURLsMap[entry.Service] = entry.URL
		}
		nomadService = v1.NewNomadService(nomadClient, standardURLsMap)
		logger.Log.Trace().Msg("Running in production mode with Nomad service")

	} else {
		nomadService = v1.NewMockNomadService()
		logger.Log.Trace().Msg("Running in local mode with mock Nomad service")
	}

	moleculeAPIService := v1.NewMoleculeAPIService(nomadService)
	moleculeAPIController := generated.NewDefaultAPIController(moleculeAPIService)

	// Add X-API-KEY authentication
	apiKey := os.Getenv("API_KEY")
	logger.Log.Trace().Msgf("API_KEY: %s", apiKey)
	if apiKey == "" {
		logger.Log.Fatal().Msg("API_KEY environment variable is required")
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

	logger.Log.Info().Msgf("Starting server on :8080")
	logger.Log.Fatal().Err(http.ListenAndServe(":8080", r))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		httpLogger := logger.Log.With().Str("request_id", requestID).Logger()
		r = r.WithContext(httpLogger.WithContext(r.Context()))
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func zerologMiddleware(next http.Handler) http.Handler {
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
	logger.Log.Debug().Msgf("Checking if route %s requires authentication", pattern)
	authenticatedRoutes := []string{
		"/v1/services/{service}/alloc-restart",
	}

	for _, route := range authenticatedRoutes {
		if route == pattern {
			logger.Log.Debug().Msgf("Route %s requires authentication", pattern)
			return true
		}
	}
	return false
}

func loadConfig(filePath string) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("failed to decode YAML file: %w", err)
	}

	return config, nil
}
