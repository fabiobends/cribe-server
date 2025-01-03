package utils

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	EncodeResponse(w, http.StatusNotFound, StandardResponse{Message: "not found"})
}

func NotAllowed(w http.ResponseWriter) {
	EncodeResponse(w, http.StatusMethodNotAllowed, StandardResponse{Message: "method not allowed"})
}
