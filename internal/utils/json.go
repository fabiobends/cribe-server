package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
)

var jsonLog = logger.NewUtilLogger("JSONUtils")

func DecodeBody[T any](r *http.Request) (T, *errors.ErrorResponse) {
	var decodedBody T

	if err := json.NewDecoder(r.Body).Decode(&decodedBody); err != nil {
		var zero T
		return zero, &errors.ErrorResponse{
			Message: errors.InvalidRequestBody,
			Details: err.Error(),
		}
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
		jsonLog.Error("Could not decode response", map[string]interface{}{
			"error":    err.Error(),
			"response": response,
		})
		panic(err) // Use panic instead of log.Fatal for utility functions
	}

	return decodedResponse
}

func SanitizeJSONString(response string) string {
	return strings.Split(response, "\n")[0]
}
