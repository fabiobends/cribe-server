package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/errors"
)

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

func DecodeResponse[T any](response string) (T, error) {
	var decodedResponse T
	err := json.Unmarshal([]byte(response), &decodedResponse)

	if err != nil {
		log.Error("Could not decode response", map[string]any{
			"error":    err.Error(),
			"response": response,
		})
		return decodedResponse, err
	}

	return decodedResponse, nil
}

func EncodeToJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Error("Could not encode to JSON", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}
	return data, nil
}

func SanitizeJSONString(response string) string {
	return strings.Split(response, "\n")[0]
}
