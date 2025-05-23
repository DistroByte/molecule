// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Molecule
 *
 * This is a simple API to list URLs and port mappings from a Nomad cluster
 *
 * API version: 1.0.0
 */

package moleculeserver

import (
	"context"
	"net/http"
)



// DefaultAPIRouter defines the required methods for binding the api requests to a responses for the DefaultAPI
// The DefaultAPIRouter implementation should parse necessary information from the http request,
// pass the data to a DefaultAPIServicer to perform the required actions, then write the service results to the http response.
type DefaultAPIRouter interface { 
	Healthcheck(http.ResponseWriter, *http.Request)
	GetURLs(http.ResponseWriter, *http.Request)
	GetServiceURLs(http.ResponseWriter, *http.Request)
	GetHostURLs(http.ResponseWriter, *http.Request)
	GetTraefikURLs(http.ResponseWriter, *http.Request)
	GetServiceStatus(http.ResponseWriter, *http.Request)
	RestartServiceAllocations(http.ResponseWriter, *http.Request)
}


// DefaultAPIServicer defines the api actions for the DefaultAPI service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can be ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type DefaultAPIServicer interface { 
	Healthcheck(context.Context) (ImplResponse, error)
	GetURLs(context.Context, bool) (ImplResponse, error)
	GetServiceURLs(context.Context) (ImplResponse, error)
	GetHostURLs(context.Context) (ImplResponse, error)
	GetTraefikURLs(context.Context) (ImplResponse, error)
	GetServiceStatus(context.Context, string) (ImplResponse, error)
	RestartServiceAllocations(context.Context, string) (ImplResponse, error)
}
