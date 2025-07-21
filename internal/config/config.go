package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/DistroByte/molecule/logger"
)

// Config represents the application configuration
type Config struct {
	Nomad struct {
		Address string `yaml:"address"`
	} `yaml:"nomad"`

	StandardURLs []StandardURL `yaml:"standard_urls"`

	ServerConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server_config"`
}

// StandardURL represents a standard URL configuration
type StandardURL struct {
	Service string `yaml:"service"`
	URL     string `yaml:"url"`
	Icon    string `yaml:"icon,omitempty"`
}

// Loader defines the interface for loading configuration
type Loader interface {
	Load(filePath string) (*Config, error)
}

// FileLoader implements configuration loading from files
type FileLoader struct{}

// NewFileLoader creates a new file-based configuration loader
func NewFileLoader() Loader {
	return &FileLoader{}
}

// Load loads configuration from the specified file path
func (f *FileLoader) Load(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			logger.Log.Error().Err(cerr).Msg("failed to close config file")
		}
	}()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Set default values
	if config.ServerConfig.Port == 0 {
		logger.Log.Warn().Msg("no port specified in server config, using default port 8080")
		config.ServerConfig.Port = 8080
	}

	logger.Log.Debug().Any("config", config).Msg("config loaded successfully")

	return &config, nil
}

// LoadFromEnvironment loads configuration from environment variables or file
func LoadFromEnvironment() (*Config, error) {
	configFilePath := os.Getenv("CONFIG_FILE")
	if configFilePath == "" {
		return nil, fmt.Errorf("CONFIG_FILE environment variable is not set")
	}

	loader := NewFileLoader()
	return loader.Load(configFilePath)
}
