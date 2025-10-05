# Cribe Server API Routes

Simple documentation about the API routes in the Cribe Go server - what they do and how they work.

## 🛣️ Routes Overview

The server has 4 main route groups that handle different functionality:

### **Auth Routes** (`/auth/*`)
- **Purpose**: Handle user authentication
- **Endpoints**: `/auth/register`, `/auth/login`, `/auth/refresh`
- **What it does**: User registration, login, and token refresh

### **Users Routes** (`/users/*`)
- **Purpose**: Manage user data (requires authentication)
- **Endpoints**: `/users`, `/users/{id}`
- **What it does**: Create, read user information

### **Status Routes** (`/status`)
- **Purpose**: Health check and system status
- **Endpoints**: `/status`
- **What it does**: Returns server health and database connection status

### **Migrations Routes** (`/migrations`)
- **Purpose**: Database migration management
- **Endpoints**: `/migrations`
- **What it does**: Preview and run database migrations

## 🔐 Auth Routes Flow

```mermaid
flowchart TD
    A[POST /auth/register] --> B[Validate User Data]
    B --> C{Valid Data?}
    C -->|No| D[Return Validation Error]
    C -->|Yes| E[Hash Password]
    E --> F[Create User in Database]
    F --> G{User Created?}
    G -->|No| H[Return Database Error]
    G -->|Yes| I[Return Success Response]

    J[POST /auth/login] --> K[Validate Credentials]
    K --> L{Valid Credentials?}
    L -->|No| M[Return Login Error]
    L -->|Yes| N[Generate Access Token]
    N --> O[Generate Refresh Token]
    O --> P[Return Tokens]

    Q[POST /auth/refresh] --> R[Validate Refresh Token]
    R --> S{Valid Token?}
    S -->|No| T[Return Unauthorized]
    S -->|Yes| U[Generate New Tokens]
    U --> V[Return New Tokens]
```

## 👥 Users Routes Flow

```mermaid
flowchart TD
    A["GET /users"] --> B[Check Authentication]
    B --> C{Authenticated?}
    C -->|No| D[Return Unauthorized]
    C -->|Yes| E[Get All Users]
    E --> F[Return Users List]

    G["GET /users/ID"] --> H[Check Authentication]
    H --> I{Authenticated?}
    I -->|No| J[Return Unauthorized]
    I -->|Yes| K[Validate User ID]
    K --> L{Valid ID?}
    L -->|No| M[Return Bad Request]
    L -->|Yes| N[Get User by ID]
    N --> O{User Found?}
    O -->|No| P[Return Not Found]
    O -->|Yes| Q[Return User Data]

    R["POST /users"] --> S[Check Authentication]
    S --> T{Authenticated?}
    T -->|No| U[Return Unauthorized]
    T -->|Yes| V[Validate User Data]
    V --> W{Valid Data?}
    W -->|No| X[Return Validation Error]
    W -->|Yes| Y[Create User]
    Y --> Z[Return Created User]
```

## 📊 Status Routes Flow

```mermaid
flowchart TD
    A[GET /status] --> B[Get Current Time]
    B --> C[Check Database Connection]
    C --> D[Format Response]
    D --> E[Return Status Info]
```

## 🔄 Migrations Routes Flow

```mermaid
flowchart TD
    A["GET /migrations"] --> B[Check Authentication]
    B --> C{Authenticated?}
    C -->|No| D[Return Unauthorized]
    C -->|Yes| E[Dry Run - Check Pending Migrations]
    E --> F[Return Migration List]

    G["POST /migrations"] --> H[Check Authentication]
    H --> I{Authenticated?}
    I -->|No| J[Return Unauthorized]
    I -->|Yes| K[Live Run - Execute Migrations]
    K --> L{Migrations Applied?}
    L -->|None| M[Return 200 OK]
    L -->|Some Applied| N[Return 201 Created]
```

## 📝 What Each Route Does

### **Auth Routes**
- **POST /auth/register**: Creates new user account with hashed password
- **POST /auth/login**: Validates credentials and returns JWT tokens
- **POST /auth/refresh**: Generates new tokens using refresh token

### **Users Routes**
- **GET /users**: Returns list of all users (authenticated users only)
- **GET /users/{id}**: Returns specific user by ID (authenticated users only)
- **POST /users**: Creates new user (authenticated users only)

### **Status Routes**
- **GET /status**: Returns server health check with database status and timestamp

### **Migrations Routes**
- **GET /migrations**: Shows what migrations would run (dry run)
- **POST /migrations**: Actually runs pending database migrations

## 🔧 Common Route Patterns

### **Request Processing Flow**
```mermaid
flowchart TD
    A[HTTP Request] --> B[Route Handler]
    B --> C[Method Check]
    C --> D{Valid Method?}
    D -->|No| E[Return Method Not Allowed]
    D -->|Yes| F[Path Validation]
    F --> G{Valid Path?}
    G -->|No| H[Return Not Found]
    G -->|Yes| I[Process Request]
    I --> J[Business Logic]
    J --> K[Database Operation]
    K --> L[Format Response]
    L --> M[Return JSON Response]
```

### **Error Handling**
```mermaid
flowchart TD
    A[Request Processing] --> B{Error Occurred?}
    B -->|No| C[Return Success Response]
    B -->|Yes| D{Error Type?}
    D -->|Validation| E[Return 400 Bad Request]
    D -->|Authentication| F[Return 401 Unauthorized]
    D -->|Not Found| G[Return 404 Not Found]
    D -->|Database| H[Return 500 Internal Server Error]
    D -->|Method Not Allowed| I[Return 405 Method Not Allowed]
```

## 🛡️ Authentication Required

Routes that require authentication (checked by middleware):
- All `/users/*` endpoints
- All `/migrations` endpoints
- Any route marked as "private" in the route configuration

Routes that are public (no authentication needed):
- `/auth/*` endpoints
- `/status` endpoint

That's it! The server routes handle authentication, user management, health checks, and database migrations in a simple, RESTful way.
