# Cribe Server Feature Flags

Simple documentation about the feature flags system in the Cribe Go server - what they do and how they work.

## 🚩 Feature Flags Overview

The server uses feature flags to control behavior in different environments, primarily for development convenience and configuration management.

### **Main Feature Flag: Development Authentication**
- **Purpose**: Allows bypassing normal JWT authentication in development
- **Environment**: Development only
- **What it does**: Auto-authenticates requests using a default user email

## 🔧 How Feature Flags Work

```mermaid
flowchart TD
    A[Request to Private Route] --> B[Check Feature Flags]
    B --> C{Dev Auth Enabled?}
    C -->|No| D[Use Normal JWT Authentication]
    C -->|Yes| E[Try Development Authentication]

    E --> F{Default Email Set?}
    F -->|No| G[Fall back to JWT Auth]
    F -->|Yes| H[Find User by Email]

    H --> I{User Found?}
    I -->|No| G
    I -->|Yes| J[Auto-authenticate User]

    D --> K[Validate JWT Token]
    G --> K
    K --> L{Valid Token?}
    L -->|No| M[Return Unauthorized]
    L -->|Yes| N[Allow Request]

    J --> N
```

## 🛠️ Development Authentication Flow

```mermaid
flowchart TD
    A[Private Route Request] --> B[MainMiddleware]
    B --> C[Check APP_ENV]
    C --> D{APP_ENV = development?}
    D -->|No| E[Skip Dev Auth]
    D -->|Yes| F[Check DEFAULT_EMAIL]

    F --> G{DEFAULT_EMAIL Set?}
    G -->|No| E
    G -->|Yes| H[Try Dev Authentication]

    H --> I[Query User by Email]
    I --> J{User Exists?}
    J -->|No| K[Dev Auth Failed]
    J -->|Yes| L[Dev Auth Success]

    L --> M[Set User ID in Context]
    M --> N[Continue Request]

    K --> E
    E --> O[Use JWT Token Auth]
    O --> P{Valid JWT?}
    P -->|No| Q[Return 401 Unauthorized]
    P -->|Yes| M
```

## 📋 Environment Variables

### **Required for Development Authentication:**
- **APP_ENV**: Must be set to `"development"`
- **DEFAULT_EMAIL**: Email of user to auto-authenticate

### **Other Configuration Variables:**
- **JWT_SECRET**: Secret key for JWT token signing
- **JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES**: Access token lifetime
- **JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS**: Refresh token lifetime
- **DATABASE_URL**: Database connection string
- **POSTGRES_DB**: Database name
- **LOG_LEVEL**: Logging level (DEBUG, INFO, WARN, ERROR)
- **LOG_DISABLE_COLORS**: Disable colored logs (`"true"` to disable)
- **LOG_DISABLE_EMOJIS**: Disable emoji logs (`"true"` to disable)
- **PORT**: Server port

## 🎯 Feature Flag Functions

### **Core Functions:**
```go
// Check if development auth is enabled
IsDevAuthEnabled() bool

// Get the current feature flags configuration
GetFeatureFlags() *FeatureFlags

// Get default email for dev auth
GetDefaultEmail() string

// Try to authenticate using default email
TryDevAuth(defaultEmail string) (int, *ErrorResponse)
```

### **Usage in Middleware:**
```mermaid
flowchart TD
    A[Request Arrives] --> B[MainMiddleware]
    B --> C{Route Private?}
    C -->|No| D[Continue Without Auth]
    C -->|Yes| E[feature_flags.IsDevAuthEnabled()]

    E --> F{Dev Auth Enabled?}
    F -->|Yes| G[feature_flags.GetDefaultEmail()]
    G --> H[feature_flags.TryDevAuth()]
    H --> I{Dev Auth Success?}
    I -->|Yes| J[Set User ID in Context]
    I -->|No| K[Use JWT Authentication]

    F -->|No| K
    K --> L[AuthMiddleware]
    L --> M{JWT Valid?}
    M -->|Yes| J
    M -->|No| N[Return 401]

    J --> O[Continue to Handler]
    D --> O
```

## 🔐 Development Authentication Benefits

### **For Developers:**
- **Skip token generation**: No need to login and get JWT tokens
- **Faster testing**: Instant access to protected routes
- **Consistent user**: Always uses the same test user
- **Easy switching**: Just set/unset environment variables

### **Security:**
- **Environment-restricted**: Only works when `APP_ENV=development`
- **Fallback protection**: Falls back to JWT auth if dev auth fails
- **Production safe**: Automatically disabled in production

## ⚙️ Configuration Examples

### **Enable Development Authentication:**
```bash
export APP_ENV=development
export DEFAULT_EMAIL=test@example.com
```

### **Disable Development Authentication:**
```bash
unset APP_ENV
# or
export APP_ENV=production
```

### **Complete Development Setup:**
```bash
export APP_ENV=development
export DEFAULT_EMAIL=dev@example.com
export LOG_LEVEL=DEBUG
export JWT_SECRET=your-secret-key
export DATABASE_URL=postgres://user:pass@localhost/db
```

## 🚦 Feature Flag States

```mermaid
stateDiagram-v2
    [*] --> CheckEnvironment
    CheckEnvironment --> DevAuthDisabled : APP_ENV != "development"
    CheckEnvironment --> CheckEmail : APP_ENV = "development"

    CheckEmail --> DevAuthDisabled : DEFAULT_EMAIL empty
    CheckEmail --> DevAuthEnabled : DEFAULT_EMAIL set

    DevAuthEnabled --> TryDevAuth : Private route request
    TryDevAuth --> DevAuthSuccess : User found
    TryDevAuth --> FallbackToJWT : User not found

    DevAuthDisabled --> UseJWTAuth : Private route request
    FallbackToJWT --> UseJWTAuth

    DevAuthSuccess --> [*]
    UseJWTAuth --> [*]
```

That's it! The feature flags system provides a simple way to enable development conveniences while maintaining security in production.
