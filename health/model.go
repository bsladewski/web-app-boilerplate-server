package health

import "time"

// healthResponse is used to format responses from the health check endpoint.
type healthResponse struct {
	Uptime      time.Duration `json:"uptime"`
	DBAvailable bool          `json:"db_available"`
}
