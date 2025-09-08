package utils

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/errors"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	EncodeResponse(w, http.StatusNotFound, errors.ErrorResponse{
		Message: errors.RouteNotFound,
		Details: "The requested resource was not found",
	})
}

func NotAllowed(w http.ResponseWriter) {
	EncodeResponse(w, http.StatusMethodNotAllowed, errors.ErrorResponse{
		Message: errors.MethodNotAllowed,
		Details: "The requested method is not allowed for this resource",
	})
}
