package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/DistroByte/molecule/logger"
)

// StaticHandler handles static file serving
type StaticHandler struct{}

// NewStaticHandler creates a new static file handler
func NewStaticHandler() *StaticHandler {
	return &StaticHandler{}
}

// ServeHome serves the main HTML page
func (h *StaticHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

// APISpecHandler handles OpenAPI specification serving
type APISpecHandler struct{}

// NewAPISpecHandler creates a new API spec handler
func NewAPISpecHandler() *APISpecHandler {
	return &APISpecHandler{}
}

// ServeSpec serves the OpenAPI specification as JSON
func (h *APISpecHandler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	specPath := "./apispec/spec/index.yaml"
	yamlData, err := os.ReadFile(specPath)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to read OpenAPI spec")
		http.Error(w, "failed to read OpenAPI spec", http.StatusInternalServerError)
		return
	}

	var jsonData map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &jsonData); err != nil {
		logger.Log.Error().Err(err).Msg("failed to parse OpenAPI spec")
		http.Error(w, "failed to parse OpenAPI spec", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jsonData); err != nil {
		logger.Log.Error().Err(err).Msg("failed to encode OpenAPI spec as JSON")
		http.Error(w, "failed to encode OpenAPI spec as JSON", http.StatusInternalServerError)
	}
}
