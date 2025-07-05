package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	generated "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Load OpenAPI spec
	spec, err := loadOpenAPISpec("apispec/spec/index.yaml")
	if err != nil {
		t.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	// Create a mock Nomad service
	nomadService := v1.NewMockNomadService()

	// Set up the API
	moleculeAPIService := v1.NewMoleculeAPIService(nomadService)
	moleculeAPIController := generated.NewDefaultAPIController(moleculeAPIService)

	r := chi.NewRouter()
	r.Get("/health", moleculeAPIController.Healthcheck)

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make a request to the /health endpoint
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Errorf("Failed to close response body: %v", cerr)
		}
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Validate the response against the OpenAPI spec
	err = validateResponse(spec, "/health", "get", resp.StatusCode, body)
	assert.NoError(t, err, "Response does not match OpenAPI spec")
}

func loadOpenAPISpec(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true // Allow external references in the spec
	spec, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the spec to ensure it's correct
	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("OpenAPI spec validation failed: %w", err)
	}

	return spec, nil
}

func validateResponse(spec *openapi3.T, path, method string, statusCode int, body []byte) error {
	// Convert the methodd to uppercase (e.g., "get" -> "GET")
	method = strings.ToUpper(method)

	// Find the operation in the spec
	operation := spec.Paths.Find(path).GetOperation(method)
	if operation == nil {
		return fmt.Errorf("no operation defined for path %s and method %s", path, method)
	}

	// Convert the status code to a string (e.g., "200")
	statusCodeStr := fmt.Sprintf("%d", statusCode)

	// Find the response in the operation
	response := operation.Responses.Value(statusCodeStr)
	if response == nil {
		return fmt.Errorf("no response defined for status code %s", statusCodeStr)
	}

	// Validate the response body against the schema
	// Validate the response body against the schema for the content type
	contentType := "application/json" // Adjust this based on your API's response content type
	mediaType := response.Value.Content[contentType]
	if mediaType == nil {
		return fmt.Errorf("no media type defined for content type %s", contentType)
	}
	if mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return fmt.Errorf("no schema defined for content type %s", contentType)
	}
	var responseBody interface{}
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	if err := mediaType.Schema.Value.VisitJSON(responseBody); err != nil {
		return fmt.Errorf("response body validation failed: %w", err)
	}

	return nil
}
