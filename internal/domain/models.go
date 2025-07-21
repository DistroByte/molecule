package domain

// ServiceInfo represents information about a service discovered in Nomad
type ServiceInfo struct {
	Name    string
	URL     string
	Icon    string
	Fetched bool
}

// AllocationData represents data extracted from Nomad allocations
type AllocationData struct {
	ServiceURLs       []ServiceInfo
	HostReservedPorts []ServiceInfo
	ServicePorts      []ServiceInfo
}

// ServiceStatus represents the status of a specific service
type ServiceStatus map[string]string

// NomadClusterInfo provides information about the Nomad cluster
type NomadClusterInfo struct {
	Address string
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host string
	Port int
}