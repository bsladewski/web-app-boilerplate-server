// Package main is the entry point for the server application.
// Environment:
//     WEB_APP_ENABLE_DEBUG_LOG
//         bool - a flag that indicates whether the application should emit
//                debug level logs.
package main

import (
	"github.com/bsladewski/web-app-boilerplate-server/env"
	"github.com/bsladewski/web-app-boilerplate-server/server"
	"github.com/sirupsen/logrus"

	_ "github.com/bsladewski/web-app-boilerplate-server/health"
	_ "github.com/bsladewski/web-app-boilerplate-server/user/delivery"
)

const (
	// enableDebugLogVariable defines the environment variable that when set to
	// true will cause the application to emit debug level logs.
	enableDebugLogVariable = "WEB_APP_ENABLE_DEBUG_LOG"
)

// main stands up the application server.
func main() {

	if env.GetBoolSafe(enableDebugLogVariable, false) {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// run the API server
	server.Run()

}
