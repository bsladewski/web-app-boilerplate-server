// Package main is the entry point for the server application.
// Environment:
//     WEB_APP_ENABLE_DEBUG_LOG
//         bool - a flag that indicates whether the application should emit
//                debug level logs.
package main

import (
	"web-app/env"
	"web-app/server"

	"github.com/sirupsen/logrus"

	_ "web-app/health"
	_ "web-app/user/delivery"
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
