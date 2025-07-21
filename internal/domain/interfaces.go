package domain

import "context"

// ServiceDiscovery defines the interface for discovering services in a Nomad cluster
type ServiceDiscovery interface {
	// GetAllServices returns all discovered services
	GetAllServices(ctx context.Context, print bool) ([]ServiceInfo, error)

	// GetServiceURLs returns only the services with URLs (typically from Traefik)
	GetServiceURLs(ctx context.Context) ([]ServiceInfo, error)

	// GetHostPorts returns services accessible via host reserved ports
	GetHostPorts(ctx context.Context) ([]ServiceInfo, error)

	// GetServicePorts returns services accessible via service ports
	GetServicePorts(ctx context.Context) ([]ServiceInfo, error)

	// GetServiceStatus returns the status of a specific service
	GetServiceStatus(ctx context.Context, serviceName string) (ServiceStatus, error)

	// RestartService restarts all allocations for a given service
	RestartService(ctx context.Context, serviceName string) error
}

// ConfigurationProvider defines the interface for loading application configuration
type ConfigurationProvider interface {
	// Load loads configuration from the specified source
	Load(source string) (*ApplicationConfig, error)
}

// ApplicationConfig represents the complete application configuration
type ApplicationConfig struct {
	NomadCluster NomadClusterInfo
	StandardURLs []ServiceInfo
	Server       ServerConfig
}
