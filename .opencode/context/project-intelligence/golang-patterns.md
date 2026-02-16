<!-- Context: project-intelligence/golang | Priority: critical | Version: 1.0 | Updated: 2026-02-16 -->

# Go Patterns

**Purpose**: Go-specific coding patterns and best practices for finsplitter.
**Last Updated**: 2026-02-16

## Directory Structure
```
project/
├── cmd/api/main.go        # Entry point
├── internal/
│   ├── config/            # Env var loading
│   ├── domain/
│   │   ├── entity/       # Business models
│   │   └── errs/         # Domain errors (sentinel)
│   ├── app/
│   │   ├── ports/        # Interfaces (Repository, UseCase)
│   │   └── usecases/     # Business logic
│   └── gateways/
│       ├── http/v1/      # Handlers (Huma v2)
│       └── postgres/     # DB (pgx, sqlc)
└── pkg/                  # Reusable libraries
```

## Key Patterns

### Functional Options
```go
type Option func(*Server)
func WithPort(port int) Option {
    return func(s *Server) { s.Port = port }
}
func NewServer(opts ...Option) *Server {
    srv := &Server{Port: 8080}
    for _, o := range opts { o(srv) }
    return srv
}
```

### Error Wrapping
```go
func readFile(name string) ([]byte, error) {
    data, err := os.ReadFile(name)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", name, err)
    }
    return data, nil
}
```

### Context Pattern
```go
func handleRequest(ctx context.Context, req *Request) {
    select {
    case <-ctx.Done(): return // Cancelled
    default: // Process
    }
}
```

## Testing
```go
// Table-driven tests
func TestAdd(t *testing.T) {
    tests := []struct{ a, b, expect int }{{1, 2, 3}}
    for _, tc := range tests {
        if got := Add(tc.a, tc.b); got != tc.expect {
            t.Errorf("Add(%d,%d)", tc.a, tc.b)
        }
    }
}

// Mock generation
//go:generate mockgen -destination=mocks/repo.go -package=mocks . Repository
```

## Security
- Validate input via Huma schema tags
- Use parameterized queries (sqlc)
- Hash passwords with bcrypt/Argon2
- Use TLS for all external communication

## 📂 Codebase References
- **Entry**: `cmd/api/main.go`
- **Domain Errors**: `internal/domain/errs/errs.go`
- **Use Cases**: `internal/app/usecases/`
- **Handlers**: `internal/gateways/http/v1/`
