# Cribe Server

Go REST API server for the Cribe Flutter app with authentication, user management, and database operations.

## 🚀 What it does

- User authentication (JWT-based login/register)
- User management (CRUD operations)
- Database migrations
- Health monitoring

## 🛠️ Prerequisites

- [Go 1.25+](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [Air](https://github.com/air-verse/air)
- [Golangci-lint](https://golangci-lint.run/)
- [Migrate](https://github.com/golang-migrate/migrate)
- [GoDotEnv](https://github.com/joho/godotenv)
- Make (usually pre-installed)

## ⚙️ Quick Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Setup Git hooks:**
   ```bash
   make setup-hooks  # Run once or when hook scripts change
   ```

3. **Start developing:**
   ```bash
   make dev
   ```

## ⚙️ Environment

Environment files are already set up. You can edit them if needed:
- `.env.development` - Development configuration
- `.env.test` - Test configuration

## 🚀 Commands

```bash
make dev          # Start with hot reload
make run          # Start without hot reload
make test         # Run tests with coverage
make build        # Build production binary
```

## 🗃️ Database

```bash
make services-up-dev     # Start dev database
make migrate-up          # Run migrations
make migrate-down        # Rollback migrations
```

## 🛣️ API Endpoints

- `POST /auth/register` - Create account
- `POST /auth/login` - Login
- `POST /auth/refresh` - Refresh tokens
- `GET /users` - List users (auth required)
- `GET /users/{id}` - Get user (auth required)
- `GET /status` - Health check

## 🔧 Development

**Feature Flags:**
```bash
export DEFAULT_EMAIL=dev@example.com  # Skip JWT authentication
export LOG_LEVEL=DEBUG  # Set logging level (DEBUG, INFO, WARN, ERROR)
```

## 📚 Documentation

- [API Routes](docs/ROUTES_DOCUMENTATION.md)
- [Feature Flags](docs/FEATURE_FLAGS_DOCUMENTATION.md)
- [Logger](docs/LOGGER_ARCHITECTURE.md)
- [AI Instructions](docs/AI_INSTRUCTIONS.md)

That's it! A Go API server with authentication and development conveniences.
