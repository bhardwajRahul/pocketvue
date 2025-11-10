package helpers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// JSONError sends a standardized JSON error response
func JSONError(e *core.RequestEvent, statusCode int, errorMsg string) error {
	return e.JSON(statusCode, ErrorResponse{
		Error: errorMsg,
	})
}

// JSONErrorWithMessage sends a JSON error response with a detailed message
func JSONErrorWithMessage(e *core.RequestEvent, statusCode int, errorMsg, message string) error {
	return e.JSON(statusCode, ErrorResponse{
		Error:   errorMsg,
		Message: message,
	})
}

// JSONBadRequest sends a 400 Bad Request error response
func JSONBadRequest(e *core.RequestEvent, errorMsg string) error {
	return JSONError(e, http.StatusBadRequest, errorMsg)
}

// JSONUnauthorized sends a 401 Unauthorized error response
func JSONUnauthorized(e *core.RequestEvent, errorMsg string) error {
	return JSONError(e, http.StatusUnauthorized, errorMsg)
}

// JSONNotFound sends a 404 Not Found error response
func JSONNotFound(e *core.RequestEvent, errorMsg string) error {
	return JSONError(e, http.StatusNotFound, errorMsg)
}

// JSONInternalServerError sends a 500 Internal Server Error response
func JSONInternalServerError(e *core.RequestEvent, errorMsg string) error {
	return JSONError(e, http.StatusInternalServerError, errorMsg)
}

// JSONSuccess sends a standardized success response
func JSONSuccess(e *core.RequestEvent, data interface{}) error {
	return e.JSON(http.StatusOK, data)
}


