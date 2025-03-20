package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	generated "github.com/DistroByte/molecule/internal/generated/go"
)

//go:generate docker run --rm -v $PWD:/spec redocly/cli lint apispec/spec/index.yaml
//go:generate docker run -u 1000:1000 --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/apispec/spec/index.yaml -g go-server -o /local/internal/generated -c /local/apispec/server-config.yaml

var (
	httpUrl = "http://zeus.internal:4646"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var nomadService v1.NomadServiceInterface

	if os.Getenv("PROD") == "true" {
		nomadClient, err := api.NewClient(&api.Config{Address: httpUrl})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Nomad client")
			return
		}
		nomadService = v1.NewNomadService(nomadClient)
	} else {
		nomadService = v1.NewMockNomadService()
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("running locally in debug mode")
	}

	moleculeAPIService := v1.NewMoleculeAPIService(nomadService)
	moleculeAPIController := generated.NewDefaultAPIController(moleculeAPIService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Serve static files from the / path
	fs := http.StripPrefix("/", http.FileServer(http.Dir("./web")))
	r.Handle("/*", fs)

	for _, route := range moleculeAPIController.Routes() {
		r.Method(route.Method, route.Pattern, route.HandlerFunc)
	}

	log.Info().Msgf("Starting server on :8080")
	log.Fatal().Err(http.ListenAndServe(":8080", r))
}
