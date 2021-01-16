package health

import (
	"net/http"
	"time"

	"web-app/cache"
	"web-app/data"
	"web-app/server"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// init binds API endpoints for checking application health.
func init() {
	server.Router().GET(healthEndpoint,
		cache.LocalCacheMiddleware(30*time.Second), healthHandler)
}

const (
	// healthEndpoint the API endpoint that checks whether the server is able to
	// complete requests.
	healthEndpoint = "/health"
)

// startTime is set when the server starts and is used to report the server
// uptime from the
var startTime = time.Now()

// healthHandler responds with basic health information about the server.
func healthHandler(c *gin.Context) {

	// check if the database is available
	dbError := data.Ping()
	if dbError != nil {
		logrus.Error(dbError)
	}

	// write health check response
	c.JSON(http.StatusOK, healthResponse{
		Uptime:      time.Now().Sub(startTime),
		DBAvailable: dbError == nil,
	})

}
