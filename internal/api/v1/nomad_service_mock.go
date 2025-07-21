package v1

import (
	generated "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/DistroByte/molecule/logger"
)

type MockNomadService struct{}

var urls = []generated.ServiceUrl{
	{
		Service: "nomad",
		Url:     "http://zeus.internal:4646",
	},
	{
		Service: "consul",
		Url:     "http://zeus.internal:8500",
	},
	{
		Service: "traefik",
		Url:     "http://hermes.internal:8081",
	},
}

func NewMockNomadService() NomadServiceInterface {
	return &MockNomadService{}
}

func (m *MockNomadService) ExtractAll(print bool) ([]generated.ServiceUrl, error) {
	logger.Log.Debug().Msg("Mock: ExtractAll called")
	return urls, nil
}

func (m *MockNomadService) ExtractURLs() ([]generated.ServiceUrl, error) {
	logger.Log.Debug().Msg("Mock: ExtractURLs called")
	return urls, nil
}

func (m *MockNomadService) ExtractHostPorts() ([]generated.ServiceUrl, error) {
	logger.Log.Debug().Msg("Mock: ExtractHostPorts called")
	return urls, nil
}

func (m *MockNomadService) ExtractServicePorts() ([]generated.ServiceUrl, error) {
	logger.Log.Debug().Msg("Mock: ExtractServicePorts called")
	return urls, nil
}

func (m *MockNomadService) GetServiceStatus(service string) (map[string]string, error) {
	logger.Log.Debug().Msg("Mock: GetServiceStatus called")
	urlMap := make(map[string]string)
	for _, url := range urls {
		if url.Service == service {
			urlMap[url.Service] = url.Url
		}
	}

	return urlMap, nil
}

func (m *MockNomadService) RestartServiceAllocations(service string) error {
	logger.Log.Debug().Msg("Mock: RestartServiceAllocations called")
	return nil
}
