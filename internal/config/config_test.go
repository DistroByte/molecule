package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoader_Load(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		checkPort   int
	}{
		{
			name: "valid config with all fields",
			configYAML: `
nomad:
  address: "http://localhost:4646"
standard_urls:
  - service: "example"
    url: "https://example.com"
    icon: "example-icon"
server_config:
  host: "localhost"
  port: 8080
`,
			expectError: false,
			checkPort:   8080,
		},
		{
			name: "valid config without port uses default",
			configYAML: `
nomad:
  address: "http://localhost:4646"
server_config:
  host: "localhost"
`,
			expectError: false,
			checkPort:   8080, // default
		},
		{
			name: "minimal config",
			configYAML: `
nomad:
  address: "http://localhost:4646"
`,
			expectError: false,
			checkPort:   8080,
		},
		{
			name: "invalid yaml",
			configYAML: `
nomad:
  address: "http://localhost:4646"
  invalid: [
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			loader := NewFileLoader()
			config, err := loader.Load(configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				return
			}

			if !tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.checkPort, config.ServerConfig.Port)
				assert.Equal(t, "http://localhost:4646", config.Nomad.Address)
			}
		})
	}
}

func TestFileLoader_Load_FileNotFound(t *testing.T) {
	loader := NewFileLoader()
	config, err := loader.Load("nonexistent-file.yaml")

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to open config file")
}

func TestLoadFromEnvironment(t *testing.T) {
	// Test with CONFIG_FILE not set
	originalEnv := os.Getenv("CONFIG_FILE")
	defer func() {
		if originalEnv != "" {
			os.Setenv("CONFIG_FILE", originalEnv)
		} else {
			os.Unsetenv("CONFIG_FILE")
		}
	}()

	os.Unsetenv("CONFIG_FILE")
	config, err := LoadFromEnvironment()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "CONFIG_FILE environment variable is not set")

	// Test with CONFIG_FILE set to valid file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	configYAML := `
nomad:
  address: "http://test:4646"
server_config:
  port: 9090
`
	err = os.WriteFile(configPath, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	os.Setenv("CONFIG_FILE", configPath)
	config, err = LoadFromEnvironment()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "http://test:4646", config.Nomad.Address)
	assert.Equal(t, 9090, config.ServerConfig.Port)
}

func TestConfig_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")
	
	// Minimal config without server port
	configYAML := `
nomad:
  address: "http://localhost:4646"
`
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewFileLoader()
	config, err := loader.Load(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.ServerConfig.Port, "should use default port when not specified")
}