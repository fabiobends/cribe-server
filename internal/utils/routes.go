package utils

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	EncodeResponse(w, http.StatusNotFound, ErrorResponse{
		StatusCode: http.StatusNotFound,
		Message:    "not found",
		Details:    []string{"The requested resource was not found"},
	})
}

func NotAllowed(w http.ResponseWriter) {
	EncodeResponse(w, http.StatusMethodNotAllowed, ErrorResponse{
		StatusCode: http.StatusMethodNotAllowed,
		Message:    "method not allowed",
		Details:    []string{"The requested method is not allowed for this resource"},
	})
}
