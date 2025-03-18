package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/nomad/api"

	v1 "github.com/DistroByte/molecule/internal/api/v1"
	generated "github.com/DistroByte/molecule/internal/generated/go"
)

//go:generate docker run --rm -v $PWD:/spec redocly/cli lint apispec/spec/index.yaml
//go:generate docker run -u 1000:1000 --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/apispec/spec/index.yaml -g go-server -o /local/internal/generated -c /local/apispec/server-config.yaml

var (
	httpUrl = "http://zeus.internal:4646"
)

func main() {
	nomadClient, err := api.NewClient(&api.Config{Address: httpUrl})
	if err != nil {
		fmt.Println(err)
		return
	}

	nomadService := v1.NewNomadService(nomadClient)
	customService := v1.NewCustomAPIService(nomadService)
	generatedController := generated.NewDefaultAPIController(customService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Serve static files from the / path
	fs := http.StripPrefix("/", http.FileServer(http.Dir("./web")))
	r.Handle("/*", fs)

	for _, route := range generatedController.Routes() {
		r.Method(route.Method, route.Pattern, route.HandlerFunc)
	}

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", r))
}
