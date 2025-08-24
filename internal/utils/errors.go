package utils

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

const (
	InvalidRequestBody         = "Invalid request body"
	RouteNotFound              = "Route not found"
	ValidationError            = "Validation error"
	DatabaseError              = "Database error"
	UserNotFound               = "User not found"
	InvalidIdParameter         = "Invalid id parameter"
	InternalError              = "Internal server error"
	InvalidAuthorizationHeader = "Invalid authorization header"
	PrivateRoute               = "Private route"
	InvalidCredentials         = "Invalid credentials" // #nosec G101
)
