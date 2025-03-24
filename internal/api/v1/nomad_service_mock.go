package v1

import "github.com/rs/zerolog/log"

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
	log.Debug().Msg("Mock: ExtractAll called")
	return urls, nil
}

func (m *MockNomadService) ExtractURLs() (map[string]string, error) {
	log.Debug().Msg("Mock: ExtractURLs called")
	return urls, nil
}

func (m *MockNomadService) ExtractHostPorts() (map[string]string, error) {
	log.Debug().Msg("Mock: ExtractHostPorts called")
	return urls, nil
}

func (m *MockNomadService) ExtractServicePorts() (map[string]string, error) {
	log.Debug().Msg("Mock: ExtractServicePorts called")
	return urls, nil
}
