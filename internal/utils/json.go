package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func DecodeBody[T any](r *http.Request) (T, *ErrorResponse) {
	var decodedBody T

	if err := json.NewDecoder(r.Body).Decode(&decodedBody); err != nil {
		var zero T
		return zero, NewErrorResponse(http.StatusBadRequest, "Invalid request body", err.Error())
	}

	return decodedBody, nil
}

func EncodeResponse(w http.ResponseWriter, statusCode int, response any) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DecodeResponse[T any](response string) T {
	var decodedResponse T
	err := json.Unmarshal([]byte(response), &decodedResponse)

	if err != nil {
		log.Fatal("Could not decode response: ", err)
	}

	return decodedResponse
}

func SanitizeJSONString(response string) string {
	return strings.Split(response, "\n")[0]
}
