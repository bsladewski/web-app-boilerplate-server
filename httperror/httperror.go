package httperror

// define generic HTTP error messages
const (
	InternalServerError = "internal server error"
)

// ErrorResponse is used to respond to an HTTP request with an error message.
type ErrorResponse struct {
	ErrorMessage string `json:"error"`
}
