package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	generated "github.com/DistroByte/molecule/internal/generated/go"
)

//go:generate docker run --rm -v $PWD:/spec redocly/cli lint apispec/spec/index.yaml
//go:generate docker run -u 1000:1000 --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/apispec/spec/index.yaml -g go-server -o /local/internal/generated -c /local/apispec/server-config.yaml

var (
	nomadUrl   = "http://zeus.internal:4646"
	traefikUrl = "http://hermes.internal:8081"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var nomadService v1.NomadServiceInterface

	if os.Getenv("PROD") == "true" {
		nomadClient, err := api.NewClient(&api.Config{Address: nomadUrl})
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
	r.Use(requestIDMiddleware)
	r.Use(zerologMiddleware)

	// Serve static files from the / path
	fs := http.StripPrefix("/", http.FileServer(http.Dir("./web")))
	r.Handle("/*", fs)

	for _, route := range moleculeAPIController.Routes() {
		r.Method(route.Method, route.Pattern, route.HandlerFunc)
	}

	log.Info().Msgf("Starting server on :8080")
	log.Fatal().Err(http.ListenAndServe(":8080", r))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		logger := log.With().Str("request_id", requestID).Logger()
		r = r.WithContext(logger.WithContext(r.Context()))
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func zerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		logger := zerolog.Ctx(r.Context())
		defer func() {
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Str("remote", r.RemoteAddr).
				Dur("duration", time.Since(start)).
				Msg("handled request")
		}()

		next.ServeHTTP(ww, r)
	})
}
