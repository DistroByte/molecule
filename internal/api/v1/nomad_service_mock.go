package v1

import (
	"github.com/DistroByte/molecule/logger"
)

type MockNomadService struct{}

var urls = map[string]string{
	"nomad":     "http://zeus.internal:4646",
	"consul":    "http://zeus.internal:8500",
	"traefik":   "http://hermes.internal:8081",
	"plausible": "https://plausible.dbyte.xyz",
}

func NewMockNomadService() NomadServiceInterface {
	return &MockNomadService{}
}

func (m *MockNomadService) ExtractAll(print bool) (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: ExtractAll called")
	return urls, nil
}

func (m *MockNomadService) ExtractURLs() (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: ExtractURLs called")
	return urls, nil
}

func (m *MockNomadService) ExtractHostPorts() (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: ExtractHostPorts called")
	return urls, nil
}

func (m *MockNomadService) ExtractServicePorts() (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: ExtractServicePorts called")
	return urls, nil
}

func (m *MockNomadService) GetServiceStatus(service string) (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: GetServiceStatus called")
	return urls, nil
}

func (m *MockNomadService) RestartServiceAllocations(service string) (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: RestartServiceAllocations called")
	return urls, nil
}
