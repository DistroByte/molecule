package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/nomad/api"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	"github.com/DistroByte/molecule/internal/config"
	generated "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/DistroByte/molecule/internal/server"
	"github.com/DistroByte/molecule/logger"
)

//go:generate docker run --rm -v $PWD:/spec redocly/cli lint apispec/spec/index.yaml
//go:generate docker run -u 1000:1000 --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/apispec/spec/index.yaml -g go-server -o /local/internal/generated -c /local/apispec/server-config.yaml

func main() {
	logger.InitLogger()
	logger.Log.Info().Msg("logger initialized")

	// Load configuration
	var cfg *config.Config
	var nomadService v1.NomadServiceInterface

	if os.Getenv("PROD") == "true" {
		var err error
		cfg, err = config.LoadFromEnvironment()
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("failed to load configuration")
		}

		if len(cfg.StandardURLs) == 0 {
			logger.Log.Warn().Msg("no standard URLs found in configuration")
		}

		// Create Nomad client
		nomadClient, err := api.NewClient(&api.Config{Address: cfg.Nomad.Address})
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to create nomad client")
			return
		}

		// Convert config URLs to generated format
		var standardURLsSlice []generated.GetUrls200ResponseInner
		for _, entry := range cfg.StandardURLs {
			standardURLsSlice = append(standardURLsSlice, generated.GetUrls200ResponseInner{
				Service: entry.Service,
				Url:     entry.URL,
				Icon:    entry.Icon,
				Fetched: false,
			})
		}
		nomadService = v1.NewNomadService(nomadClient, standardURLsSlice)
	} else {
		nomadService = v1.NewMockNomadService()
		cfg = &config.Config{} // Default config for dev mode
		cfg.ServerConfig.Port = 8080
	}

	// Validate API key
	apiKey := os.Getenv("API_KEY")
	logger.Log.Trace().Msgf("API_KEY: %s", apiKey)
	if apiKey == "" {
		logger.Log.Fatal().Msg("API_KEY environment variable is required")
	}

	// Create services
	moleculeAPIService := v1.NewMoleculeAPIService(nomadService)
	moleculeAPIController := generated.NewDefaultAPIController(moleculeAPIService)

	// Create and configure server
	srv := server.New(cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	r := srv.Router()

	// Setup routes
	setupRoutes(r, moleculeAPIController, apiKey)

	// Start server
	logger.Log.Fatal().Err(srv.Start()).Msg("server failed to start")
}

// setupRoutes configures all application routes
func setupRoutes(r chi.Router, moleculeAPIController *generated.DefaultAPIController, apiKey string) {
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

	// Setup API routes with authentication
	apiRouter := chi.NewRouter()
	apiRouter.Use(server.APIKeyAuthMiddleware(apiKey))
	
	for _, route := range moleculeAPIController.Routes() {
		if server.RequiresAuth(route.Pattern) {
			apiRouter.Method(route.Method, route.Pattern, route.HandlerFunc)
		} else {
			r.Method(route.Method, route.Pattern, route.HandlerFunc)
		}
	}
	r.Mount("/", apiRouter)
}
