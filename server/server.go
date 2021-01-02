// Package server exposes a server router than can be used to bind API endpoints
// and provides functions for managing the server.
//
// Environment:
//     WEB_APP_PORT
//         int - the port on which we listen for incoming requests.
//     WEB_APP_CERT
//         string - the path to the certificate used for TLS encryption.
//     WEB_APP_KEY:
//         string - the path to the key used for TLS encryption.
//     WEB_APP_CORS_ALLOW_ORIGINS
//         string - a comma separated list of origins a cross-domain request
//                  can be executed from.
//                  Default: *
//     WEB_APP_CORS_ALLOW_METHODS
//         string - a comma separated list of HTTP methods a client is allowed
//                  to use in a cross-domain request.
//                  Default: POST, GET, PUT, PATCH, DELETE
//     WEB_APP_CORS_ALLOW_HEADERS
//         string - a comma separated list of headers a client is allowed to use
//                  in a cross-domain request.
//                  Default: Accept, Content-Type, Content-Length,
//                           Accept-Encoding, X-CSRF-Token, Authorization,
//                           Origin, Cache-Control, X-Requested-With
//     WEB_APP_CORS_ALLOW_CREDENTIALS
//         bool - a flag that indicates whether a cross-domain request may
//                include user credentials.
//                Default: true
//     WEB_APP_CORS_EXPOSE_HEADERS
//         string - a comma separated list of headers the server may expose in
//                  responses to cross-domain requests.
//                  Default: X-Requested-With, X-Total-Records
//     WEB_APP_CORS_MAX_AGE
//         int - the number of seconds a preflight response may be cached.
//               Default: 600
//     WEB_APP_CLIENT_HOST
//         string - the host that is used to server the application front-end.
package server

import (
	"fmt"
	"regexp"
	"time"

	"github.com/bsladewski/web-app-boilerplate-server/env"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// init initializes the application server router.
func init() {

	if router != nil {
		return
	}

	r := regexp.MustCompile("\\s*,\\s*")

	// parse CORS settings from environment
	allowOrigins = r.Split(env.GetStringSafe(allowOriginsVariable,
		"*"), -1)
	allowMethods = r.Split(env.GetStringSafe(allowMethodsVariable,
		"POST,GET,PUT,PATCH,DELETE"), -1)
	allowHeaders = r.Split(env.GetStringSafe(allowHeadersVariable,
		"Accept,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Authorization,Origin,Cache-Control,X-Requested-With"), -1)
	allowCredentials = env.GetBoolSafe(allowCredentialsVariable, true)
	exposeHeaders = r.Split(env.GetStringSafe(exposeHeadersVariable,
		"X-Requested-With,X-Total-Records"), -1)
	preflightMaxAge = env.GetIntSafe(preflightMaxAgeVariable, 600)

	// get client host
	clientHost = env.MustGetString(clientHostVariable)

	// initialize application server router
	router = gin.Default()

	// initialize CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: allowCredentials,
		ExposeHeaders:    exposeHeaders,
		MaxAge:           time.Duration(preflightMaxAge) * time.Second,
	}))
}

const (
	// portVariable defines the environment variable for the server port.
	portVariable = "WEB_APP_PORT"
	// tlsCertVariable defines the environment variable for the TLS certificate.
	// If set the server will be run using TLS encryption.
	tlsCertVariable = "WEB_APP_CERT"
	// tlsKeyVariable defines the environment variable for the TLS key. If set
	// the server will run using TLS encryption.
	tlsKeyVariable = "WEB_APP_KEY"
	// httpDefaultPort the default port when running the server without TLS
	// encryption and no explicit port.
	httpDefaultPort = 80
	// httpsDefaultPort the default port when running the server with TLS
	// encryption and no explicit port.
	httpsDefaultPort = 443
	// allowOriginsVariable defines the environment variable for the allow
	// origins CORS policy.
	allowOriginsVariable = "WEB_APP_CORS_ALLOW_ORIGINS"
	// allowMethodsVariable defines the environment variable for the allow
	// methods CORS policy.
	allowMethodsVariable = "WEB_APP_CORS_ALLOW_METHODS"
	// allowHeadersVariable defines the environment variable for the allow
	// headers CORS policy.
	allowHeadersVariable = "WEB_APP_CORS_ALLOW_HEADERS"
	// allowCredentialsVariable defines the environment variable for the allow
	// credentials CORS policy.
	allowCredentialsVariable = "WEB_APP_CORS_ALLOW_CREDENTIALS"
	// exposeHeadersVariable defines the environment variable for the expose
	// headers CORS policy.
	exposeHeadersVariable = "WEB_APP_CORS_EXPOSE_HEADERS"
	// preflightMaxAgeVariable defines the environment variable for the max age
	// of cached preflight reponse.
	preflightMaxAgeVariable = "WEB_APP_CORS_MAX_AGE"
	// clientHostVariable defines the environment variable for the host that is
	// used to serve the application front-end.
	clientHostVariable = "WEB_APP_CLIENT_HOST"
)

// router is used to bind API endpoints.
var router *gin.Engine

// allowOrigins determines which origins may execute a cross-domain request.
var allowOrigins []string

// allowMethods determines which HTTP methods a client may use in a cross-domain
// request.
var allowMethods []string

// allowHeaders determines which headers may be supplied in a cross-domain
// request.
var allowHeaders []string

// allowCredentials determines whether a cross-domain request may include user
// credentials.
var allowCredentials bool

// exposeHeaders determines which headers the server may expose in responses to
// cross-domain requests.
var exposeHeaders []string

// preflightMaxAge determines how long in seconds we may cache a response to a
// preflight request.
var preflightMaxAge int

// clientHost stores the host that serves the application front-end for use in
// formatting links.
var clientHost string

// Router retrieves the application server router which can be used to bind
// handler functions to API endpoints.
func Router() *gin.Engine {
	return router
}

// ClientHost retrieves the client host.
func ClientHost() string {
	return clientHost
}

// Run starts the application server. Returns when the server is terminated.
func Run() {

	cert, key := env.GetString(tlsCertVariable), env.GetString(tlsKeyVariable)

	// check if we should be running the server using TLS encryption
	if cert != "" || key != "" {

		// run the server using HTTPS
		port := env.GetIntSafe(portVariable, httpsDefaultPort)

		logrus.Infof("starting HTTPS server on port %d", port)
		logrus.Error(router.RunTLS(
			fmt.Sprintf(":%d", port),
			cert, key,
		))

	} else {

		// run the server using HTTP
		port := env.GetIntSafe(portVariable, httpDefaultPort)

		logrus.Infof("starting HTTP server on port %d", port)
		logrus.Error(router.Run(
			fmt.Sprintf(":%d", port),
		))

	}

}
