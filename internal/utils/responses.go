package utils

type StandardResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	StatusCode int      `json:"status_code"`
	Message    string   `json:"message"`
	Details    []string `json:"details"`
}
