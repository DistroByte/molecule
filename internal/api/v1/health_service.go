package v1

import (
	"context"
	"net/http"

	openapi "github.com/DistroByte/molecule/internal/generated/go"
)

func (s *MoleculeAPIService) Healthcheck(ctx context.Context) (openapi.ImplResponse, error) {
	return openapi.Response(http.StatusOK, "OK"), nil
}
