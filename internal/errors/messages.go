package errors

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

// HTTP and General Errors
const (
	InvalidRequestBody  = "Invalid request body"
	RouteNotFound       = "Route not found"
	MethodNotAllowed    = "Method not allowed"
	ValidationError     = "Validation error"
	InvalidIdParameter  = "Invalid id parameter"
	InternalServerError = "Internal server error"
)

// Database Errors
const (
	DatabaseError    = "Database error"
	DatabaseNotFound = "Database record not found"
)

// External API Errors
const (
	ExternalAPIError = "External API error"
)

// Authentication and Authorization Errors
const (
	InvalidAuthorizationHeader = "Invalid authorization header"
	InvalidCredentials         = "Invalid credentials" // #nosec G101
	PrivateRoute               = "Private route"
	Unauthorized               = "Unauthorized"
)

// User-related Errors
const (
	UserNotFound      = "User not found"
	UserAlreadyExists = "User already exists"
)

// Development and Feature Flag Errors
const (
	DevAuthNotEnabled = "Development authentication not enabled"
	DevAuthFailed     = "Development authentication failed"
)

// Migration Errors
const (
	MigrationPathError = "Failed to get migrations path"
)
