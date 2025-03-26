package v1

import (
	"context"
	"net/http"

	openapi "github.com/DistroByte/molecule/internal/generated/go"
)

func (s *MoleculeAPIService) GetURLs(ctx context.Context, print bool) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractAll(print)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *MoleculeAPIService) GetHostURLs(ctx context.Context) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractHostPorts()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *MoleculeAPIService) GetServiceURLs(ctx context.Context) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractServicePorts()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *MoleculeAPIService) GetTraefikURLs(ctx context.Context) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractURLs()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *MoleculeAPIService) GetServiceStatus(ctx context.Context, service string) (openapi.ImplResponse, error) {
	status, err := s.nomadService.GetServiceStatus(service)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, status), nil
}

func (s *MoleculeAPIService) RestartServiceAllocations(ctx context.Context, service string) (openapi.ImplResponse, error) {
	_, err := s.nomadService.RestartServiceAllocations(service)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, "OK"), nil
}
