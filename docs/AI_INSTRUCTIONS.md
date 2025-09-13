# AI Instructions

## Role & Responsibilities
- You are a Go backend development assistant focused on performance, security, and best practices
- Follow clean architecture principles and maintain existing code patterns
- Use proper error handling and logging throughout all code
- Prioritize code quality, type safety, and maintainability over quick solutions

## Project Structure
```
cmd/
├── app/            # Application entry point
│   ├── main.go     # Server startup and configuration
│   └── main_test.go
internal/
├── core/           # Core server logic
│   ├── logger/     # Logging system (interface, console logger, service)
│   └── router.go   # HTTP routing and server setup
├── errors/         # Error handling and messages
├── feature_flags/  # Feature flag management
├── middlewares/    # HTTP middlewares (auth, CORS, etc.)
├── routes/         # Feature-based route handlers
│   ├── auth/       # Authentication endpoints
│   ├── migrations/ # Database migration endpoints
│   ├── status/     # Health check endpoints
│   └── users/      # User management endpoints
└── utils/          # Shared utilities and helpers
infra/
├── migrations/     # Database migration files
```

## Code Guidelines

### Architecture Rules
- **Handlers**: HTTP request/response handling, validation, routing
- **Services**: Business logic, orchestration between repositories
- **Repositories**: Data access layer, database operations
- **Models**: Data structures, validation, serialization
- **Middlewares**: Cross-cutting concerns (auth, logging, CORS)
- **Utils**: Shared utilities, helper functions

### Go Conventions
- Use `snake_case` for file names (e.g., `user_service.go`)
- Use `PascalCase` for exported types and functions
- Use `camelCase` for unexported variables and functions
- Follow standard Go project layout conventions
- Use interfaces for testability and dependency injection

### Error Handling
- Always handle errors explicitly, never ignore them
- Use custom error types in `/internal/errors/` package
- Return `*errors.ErrorResponse` for business logic errors
- Log errors with appropriate context before returning
- Use structured error responses with message and details

### Logging System
- Use contextual loggers: `NewHandlerLogger()`, `NewServiceLogger()`, `NewRepositoryLogger()`, etc.
- Logger automatically masks sensitive data (emails, passwords, JWT tokens)
- Environment-controlled log levels via `LOG_LEVEL` (DEBUG/INFO/WARN/ERROR/NONE)
- ANSI colors and emojis for visual distinction in development
- Platform-agnostic interface ready for structured logging, Sentry, etc.

### HTTP Layer
- Validate all incoming request data using struct validation tags
- Use `utils.ValidateStruct()` for comprehensive validation
- Use appropriate HTTP status codes
- Return consistent JSON response format
- Handle method routing in handlers
- Use middlewares for cross-cutting concerns

### Validation Guidelines
- Use `github.com/go-playground/validator/v10` for all data validation
- Add validation tags to struct fields (e.g., `validate:"required,email,min=8"`)
- Use `utils.ValidateStruct()` instead of manual validation
- Available validation functions: `IsValidEmail()`, `IsValidPassword()`, `IsValidName()`
- Never use hardcoded validation like `strings.Contains(email, "@")`

## Testing & Quality

### Testing Strategy
- **Handlers** → integration tests with mock services
- **Services** → unit tests with mock repositories
- **Repositories** → integration tests with test database
- **Utils** → unit tests for pure functions

### Code Quality
- Run `make lint` before committing (golangci-lint)
- Run `make format` to ensure consistent formatting
- Run `make typecheck` for static analysis
- Use `make test` for full test suite with test database
- Remove anything you’ve created that’s not used in any file
- Follow `pre-commit` and `pre-push` hooks

### Database
- Use migrations for all schema changes (`make migrate-create`)
- Use parameterized queries to prevent SQL injection
- Handle database connection errors gracefully
- Use transactions for multi-operation consistency

## Development Workflow

### Environment Setup
- Use `make dev` for hot reload development
- Use `make dev-debug` for debugging with Delve
- Environment configuration in `.env.development` and `.env.test`
- Configure logging via environment variables:
  - `LOG_LEVEL`: DEBUG/INFO/WARN/ERROR/NONE (default: INFO)
  - `LOG_DISABLE_COLORS`: true/false (default: false)
  - `LOG_DISABLE_EMOJIS`: true/false (default: false)

### Dependencies
- Keep `go.mod` minimal, only add necessary dependencies
- Use `make update-deps` to update dependencies
- Prefer standard library when possible
- Current stack: pgx/v5, golang-jwt, golang-migrate, bcrypt, validator/v10

## Security Guidelines
- Hash passwords using bcrypt with appropriate cost
- Use JWT tokens for authentication with proper expiration
- Validate and sanitize all user inputs
- Use parameterized queries for database operations
- Implement proper CORS and security headers
- Log security-relevant events (auth failures, etc.)

## Feature Development
- Add new routes in `/internal/routes/[feature]/`
- Create handler, service, repository, and model files
- Add appropriate contextual logging to all new code
- Add corresponding tests for each layer
- Update router in `/internal/core/router.go`
- Add database migrations if needed

### Logger Usage Examples
```go
// In handlers
logger := logger.NewHandlerLogger("AuthHandler")
logger.Info("Processing login request", map[string]interface{}{"email": user.Email})

// In services
logger := logger.NewServiceLogger("AuthService")
logger.Error("Failed to generate token", map[string]interface{}{"userID": userID, "error": err.Error()})

// In repositories
logger := logger.NewRepositoryLogger("UserRepository")
logger.Debug("Executing user query", map[string]interface{}{"query": query})
```

## Restrictions
- Don't use `panic()` in production code, handle errors gracefully
- Don't ignore errors with `_` assignment
- Don't use global variables for application state
- Don't run database operations in handlers directly
- Don't run `go build ./...` or something similar to see if something broken prefer `make typecheck` for this
- Keep all summaries of changes straightforward and concise
