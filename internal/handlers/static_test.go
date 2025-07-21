package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticHandler_ServeHome(t *testing.T) {
	// Create a temporary web directory with index.html
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	webDir := "./web"
	err = os.MkdirAll(webDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create web directory: %v", err)
	}

	indexContent := "<html><body><h1>Hello, World!</h1></body></html>"
	indexPath := filepath.Join(webDir, "index.html")
	err = os.WriteFile(indexPath, []byte(indexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}

	handler := NewStaticHandler()

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHome(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Hello, World!")
}

func TestStaticHandler_ServeHome_FileNotFound(t *testing.T) {
	// Create a temporary directory without index.html
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	handler := NewStaticHandler()

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHome(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestAPISpecHandler_ServeSpec(t *testing.T) {
	// Create a temporary directory with OpenAPI spec
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	apispecDir := "./apispec/spec"
	err = os.MkdirAll(apispecDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apispec directory: %v", err)
	}

	specContent := `
openapi: 3.0.0
info:
  title: Molecule API
  version: 1.0.0
paths:
  /health:
    get:
      summary: Health check
      responses:
        '200':
          description: OK
`
	specPath := filepath.Join(apispecDir, "index.yaml")
	err = os.WriteFile(specPath, []byte(specContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	handler := NewAPISpecHandler()

	req := httptest.NewRequest("GET", "/spec", nil)
	recorder := httptest.NewRecorder()

	handler.ServeSpec(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "Molecule API")
	assert.Contains(t, recorder.Body.String(), "health")
}

func TestAPISpecHandler_ServeSpec_FileNotFound(t *testing.T) {
	// Create a temporary directory without spec file
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	handler := NewAPISpecHandler()

	req := httptest.NewRequest("GET", "/spec", nil)
	recorder := httptest.NewRecorder()

	handler.ServeSpec(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "failed to read OpenAPI spec")
}

func TestAPISpecHandler_ServeSpec_InvalidYAML(t *testing.T) {
	// Create a temporary directory with invalid YAML spec
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	apispecDir := "./apispec/spec"
	err = os.MkdirAll(apispecDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create apispec directory: %v", err)
	}

	invalidSpecContent := `
openapi: 3.0.0
info:
  title: Molecule API
  version: [ invalid yaml
`
	specPath := filepath.Join(apispecDir, "index.yaml")
	err = os.WriteFile(specPath, []byte(invalidSpecContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	handler := NewAPISpecHandler()

	req := httptest.NewRequest("GET", "/spec", nil)
	recorder := httptest.NewRecorder()

	handler.ServeSpec(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "failed to parse OpenAPI spec")
}
