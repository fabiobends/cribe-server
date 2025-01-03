package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

func EncodeResponse(w http.ResponseWriter, statusCode int, response any) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SanitizeJSONString(response string) string {
	return strings.Split(response, "\n")[0]
}
