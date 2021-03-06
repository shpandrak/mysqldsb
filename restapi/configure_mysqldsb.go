package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	graceful "github.com/tylerb/graceful"

	"ocopea/mysqldsb/restapi/operations"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target .. --name mysqldsb --spec ../dsb-swagger.yaml

func configureFlags(api *operations.MysqldsbAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.MysqldsbAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// s.api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.DsbWebCopyServiceInstanceHandler = dsb_web.CopyServiceInstanceHandlerFunc(func(params dsb_web.CopyServiceInstanceParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.CopyServiceInstance has not yet been implemented")
	})
	api.DsbWebCreateServiceInstanceHandler = dsb_web.CreateServiceInstanceHandlerFunc(func(params dsb_web.CreateServiceInstanceParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.CreateServiceInstance has not yet been implemented")
	})
	api.DsbWebDeleteServiceInstanceHandler = dsb_web.DeleteServiceInstanceHandlerFunc(func(params dsb_web.DeleteServiceInstanceParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.DeleteServiceInstance has not yet been implemented")
	})
	api.DsbWebGetDSBIconHandler = dsb_web.GetDSBIconHandlerFunc(func(params dsb_web.GetDSBIconParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.GetDSBIcon has not yet been implemented")
	})
	api.DsbWebGetDSBInfoHandler = dsb_web.GetDSBInfoHandlerFunc(func(params dsb_web.GetDSBInfoParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.GetDSBInfo has not yet been implemented")
	})
	api.DsbWebGetServiceInstanceHandler = dsb_web.GetServiceInstanceHandlerFunc(func(params dsb_web.GetServiceInstanceParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.GetServiceInstance has not yet been implemented")
	})
	api.DsbWebGetServiceInstancesHandler = dsb_web.GetServiceInstancesHandlerFunc(func(params dsb_web.GetServiceInstancesParams) middleware.Responder {
		return middleware.NotImplemented("operation dsb_web.GetServiceInstances has not yet been implemented")
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
