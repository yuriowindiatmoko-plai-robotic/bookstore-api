# AGENTS.md

## Project: Bookstore REST API

### Stack
- Go 1.23+, Gin framework, GORM ORM
- PostgreSQL 16 (Podman, port 5432)
- JWT authentication (golang-jwt/v5)

### Architecture
- `main.go` — entry point, server setup
- `config/` — database connection, env loading
- `models/` — GORM models (Book, User)
- `handlers/` — Gin handler functions
- `middleware/` — JWT auth middleware
- `routes/` — route definitions

### Conventions
- Error responses: `{"error": "message"}` with proper HTTP status
- Success responses: `{"data": {...}}` or `{"data": [...], "meta": {...}}`
- Pagination: `?page=1&limit=10`, default limit 10
- Use GORM auto-migrate (development only)
- Password hashing: bcrypt via golang.org/x/crypto
- JWT in Authorization header: `Bearer <token>`

### Do
- Proper error handling (don't ignore errors)
- Use pointer receivers for methods
- Group routes under /api/v1
- Return proper HTTP status codes (201 created, 404 not found, etc)

### Don't
- Don't use `fmt.Println` for logging (use Gin's logger)
- Don't store passwords as plaintext
- Don't skip input validation
- Don't return database errors to client (security risk)