package v1

import (
	"context"
	"net/http"

	openapi "github.com/DistroByte/molecule/internal/generated/go"
)

type CustomAPIService struct {
	nomadService *NomadService
}

func NewCustomAPIService(nomadService *NomadService) *CustomAPIService {
	return &CustomAPIService{nomadService: nomadService}
}

func (s *CustomAPIService) GetURLs(ctx context.Context, print bool) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractAll(print)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *CustomAPIService) GetHostURLs(ctx context.Context) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractHostPorts()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *CustomAPIService) GetServiceURLs(ctx context.Context) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractServicePorts()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}

func (s *CustomAPIService) GetTraefikURLs(ctx context.Context) (openapi.ImplResponse, error) {
	urls, err := s.nomadService.ExtractURLs()
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, err.Error()), nil
	}

	// Return the response
	return openapi.Response(http.StatusOK, urls), nil
}
