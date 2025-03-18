package v1

import (
	openapi "github.com/DistroByte/molecule/internal/generated/go"
	"github.com/go-chi/chi/v5"
)

type APIController struct {
	service openapi.DefaultAPIServicer
}

func NewCustomAPIController(service openapi.DefaultAPIServicer) *APIController {
	return &APIController{service: service}
}

func (c *APIController) Routes() chi.Router {
	r := chi.NewRouter()
	return r
}
