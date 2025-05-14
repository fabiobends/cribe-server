package utils

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	EncodeResponse(w, http.StatusNotFound, ErrorResponse{
		Message: "Not found",
		Details: "The requested resource was not found",
	})
}

func NotAllowed(w http.ResponseWriter) {
	EncodeResponse(w, http.StatusMethodNotAllowed, ErrorResponse{
		Message: "Method not allowed",
		Details: "The requested method is not allowed for this resource",
	})
}
