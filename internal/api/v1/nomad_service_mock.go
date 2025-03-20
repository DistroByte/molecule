package v1

import "github.com/rs/zerolog/log"

type MockNomadService struct{}

func NewMockNomadService() NomadServiceInterface {
	return &MockNomadService{}
}

func (m *MockNomadService) ExtractAll(print bool) (map[string]string, error) {
	return map[string]string{
			"nomad":    "http://zeus.internal:4646",
			"consul":   "http://zeus.internal:8500",
			"traefik":  "http://zeus.internal:8081",
			"plausible": "https://plausible.dbyte.xyz",
	}, nil
}

func (m *MockNomadService) ExtractURLs() (map[string]string, error) {
	log.Debug().Msg("Mock: ExtractURLs called")
	return map[string]string{
		"mock-service": "http://mock-service.local",
	}, nil
}

func (m *MockNomadService) ExtractHostPorts() (map[string]string, error) {
	log.Debug().Msg("Mock: ExtractHostPorts called")
	return map[string]string{
		"mock-host-port": "127.0.0.1:1234",
	}, nil
}

func (m *MockNomadService) ExtractServicePorts() (map[string]string, error) {
	log.Debug().Msg("Mock: ExtractServicePorts called")
	return map[string]string{
		"mock-service-port": "127.0.0.1:5678",
	}, nil
}
