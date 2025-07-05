package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog"

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

	ServerConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server_config"`
}

var config Config

func main() {
	logger.InitLogger()

	logger.Log.Info().Msg("logger initialized")

	var nomadService v1.NomadServiceInterface

	if os.Getenv("PROD") == "true" {
		configFilePath := os.Getenv("CONFIG_FILE")
		var err error

		if configFilePath == "" {
			logger.Log.Fatal().Msg("CONFIG_FILE environment variable is not set - not loading config")
		} else {
			config, err = loadConfig(configFilePath)

			if err != nil {
				logger.Log.Error().Err(err).Msg("failed to load config file")
				return
			}
		}

		if len(config.StandardURLs) == 0 {
			logger.Log.Warn().Msg("no standard URLs found in YAML file")
		}

		nomadClient, err := api.NewClient(&api.Config{Address: config.Nomad.Address})
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to create api client")
			return
		}

		// make an array of URLInfo from the standard URLs
		var standardURLsSlice []generated.GetUrls200ResponseInner
		for _, entry := range config.StandardURLs {
			standardURLsSlice = append(standardURLsSlice, generated.GetUrls200ResponseInner{
				Service: entry.Service,
				Url:     entry.URL,
				Fetched: false,
			})
		}
		nomadService = v1.NewNomadService(nomadClient, standardURLsSlice)
	} else {
		nomadService = v1.NewMockNomadService()
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
			http.Error(w, "failed to read OpenAPI spec", http.StatusInternalServerError)
			return
		}

		var jsonData map[string]interface{}
		if err := yaml.Unmarshal(yamlData, &jsonData); err != nil {
			http.Error(w, "failed to parse OpenAPI spec", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(jsonData); err != nil {
			http.Error(w, "failed to encode OpenAPI spec as JSON", http.StatusInternalServerError)
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

	// if no host specified, use "localhost" in the log message
	var serverHost string
	if config.ServerConfig.Host == "" {
		serverHost = "localhost"
	} else {
		serverHost = config.ServerConfig.Host
	}

	var serverPort int
	if config.ServerConfig.Port == 0 {
		logger.Log.Warn().Msg("no port specified in server config, using default port 8080")
		serverPort = 8080
	} else {
		serverPort = config.ServerConfig.Port
	}

	logger.Log.Info().Msgf("starting server on http://%s:%d", serverHost, serverPort)
	logger.Log.Fatal().Err(http.ListenAndServe(fmt.Sprintf("%s:%d", config.ServerConfig.Host, serverPort), r)).Msg("Server failed to start")
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

func loadConfig(filePath string) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, fmt.Errorf("failed to decode YAML file: %w", err)
	}

	// Set default values for server config if not provided
	if config.ServerConfig.Port == 0 {
		logger.Log.Warn().Msg("no port specified in server config, using default port 8080")
		config.ServerConfig.Port = 8080
	}

	logger.Log.Debug().Any("config", config).Msg("config loaded")

	return config, nil
}
