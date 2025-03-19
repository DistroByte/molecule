package v1

import (
	"context"
	"net/http"

	openapi "github.com/DistroByte/molecule/internal/generated/go"
)

type MoleculeAPIService struct {
	nomadService *NomadService
}

func NewMoleculeAPIService(nomadService *NomadService) *MoleculeAPIService {
	return &MoleculeAPIService{nomadService: nomadService}
}

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

func (s *MoleculeAPIService) Healthcheck(ctx context.Context) (openapi.ImplResponse, error) {
	return openapi.Response(http.StatusOK, "OK"), nil
}
