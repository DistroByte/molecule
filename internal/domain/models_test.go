package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceInfo(t *testing.T) {
	service := ServiceInfo{
		Name:    "test-service",
		URL:     "https://test.example.com",
		Icon:    "test-icon",
		Fetched: true,
	}

	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "https://test.example.com", service.URL)
	assert.Equal(t, "test-icon", service.Icon)
	assert.True(t, service.Fetched)
}

func TestAllocationData(t *testing.T) {
	data := &AllocationData{
		ServiceURLs: []ServiceInfo{
			{Name: "web", URL: "https://web.example.com", Fetched: true},
		},
		HostReservedPorts: []ServiceInfo{
			{Name: "api", URL: "http://localhost:8080", Fetched: true},
		},
		ServicePorts: []ServiceInfo{
			{Name: "db", URL: "http://localhost:5432", Fetched: true},
		},
	}

	assert.Len(t, data.ServiceURLs, 1)
	assert.Len(t, data.HostReservedPorts, 1)
	assert.Len(t, data.ServicePorts, 1)

	assert.Equal(t, "web", data.ServiceURLs[0].Name)
	assert.Equal(t, "api", data.HostReservedPorts[0].Name)
	assert.Equal(t, "db", data.ServicePorts[0].Name)
}

func TestServiceStatus(t *testing.T) {
	status := ServiceStatus{
		"service1": "https://service1.example.com",
		"service2": "http://localhost:8080",
	}

	assert.Equal(t, "https://service1.example.com", status["service1"])
	assert.Equal(t, "http://localhost:8080", status["service2"])
	assert.Equal(t, "", status["nonexistent"])
}

func TestNomadClusterInfo(t *testing.T) {
	cluster := NomadClusterInfo{
		Address: "http://localhost:4646",
	}

	assert.Equal(t, "http://localhost:4646", cluster.Address)
}

func TestServerConfig(t *testing.T) {
	config := ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 8080, config.Port)
}

func TestApplicationConfig(t *testing.T) {
	config := &ApplicationConfig{
		NomadCluster: NomadClusterInfo{
			Address: "http://localhost:4646",
		},
		StandardURLs: []ServiceInfo{
			{Name: "static", URL: "https://static.example.com"},
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 9090,
		},
	}

	assert.Equal(t, "http://localhost:4646", config.NomadCluster.Address)
	assert.Len(t, config.StandardURLs, 1)
	assert.Equal(t, "static", config.StandardURLs[0].Name)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)
}
