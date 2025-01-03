package utils

type StandardResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string   `json:"message"`
	Details []string `json:"details"`
}
