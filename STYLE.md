# Project Structure Style Guide

A blueprint for small "api, database, service" style applications with CLI tooling.

## Directory Layout

```
project/
├── cmd/project/        # CLI entry point and commands
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── service/        # Business logic with store interfaces
│   ├── database/       # Persistence implementations
│   └── util/           # Shared utilities
├── init/               # Deployment configs (systemd, etc.)
└── scripts/            # Build and deployment automation
```

## Layer Architecture

Follow strict dependency flow: **API → Service → Database**

### Service Layer (internal/service/)

Define store interfaces in the service layer, not the database layer.

```go
// 1. Define interface for data operations
type LedgerStore interface {
    GetSnapshot(ledger string, since, until int64) (*Snapshot, error)
    InsertTransaction(id, ledger string, amount int) error
}

// 2. Module-level variable for DI
var ledgerStore LedgerStore

// 3. Setter for dependency injection
func SetLedgerStore(s LedgerStore) {
    ledgerStore = s
}

// 4. Business logic checks for nil store
func GetSnapshot(ledger string, since, until time.Time) (*Snapshot, error) {
    if ledgerStore == nil {
        return nil, ErrNoLedgerStore
    }
    // Business logic here
    return ledgerStore.GetSnapshot(ledger, since.Unix(), until.Unix())
}
```

**Naming**: `XxxStore` interface, `SetXxxStore()` setter, `ErrNoXxxStore` sentinel error.

### Database Layer (internal/database/)

Implement service interfaces with concrete database operations.

```go
// Implementation struct (typically empty)
type DBLedgerStore struct{}

// Constructor returns implementation
func NewLedgerStore() DBLedgerStore {
    return DBLedgerStore{}
}

// Implement interface methods
func (DBLedgerStore) GetSnapshot(ledger string, since, until int64) (*service.Snapshot, error) {
    // Database queries here
}
```

**Pattern**: Use prepared statements, parameterized queries (`?1`, `?2`), and upserts where appropriate.

### API Layer (internal/api/)

Build modular routers with middleware composition.

```go
// Router builder per resource
func buildLedgerRouter(r *mux.Router) {
    r.HandleFunc("/{id}", withCORS(handleGetLedger)).Methods("GET", "OPTIONS")
    r.HandleFunc("/{id}/tx", withAuth(handlePostTransaction)).Methods("POST")
}

// Compose into main router
func BuildRouter(r *mux.Router) {
    buildHealthRouter(r.PathPrefix("/health").Subrouter())
    buildLedgerRouter(r.PathPrefix("/ledger").Subrouter())
}
```

**Standard response format**:
```go
type APIResponse struct {
    Error *APIError `json:"error"`
    Data  any       `json:"data"`
}
```

**Helper functions**: `writeData(w, status, data)`, `writeError(w, status, message)`, `writeJSON(w, status, v)`

## Initialization Pattern

Initialize layers in dependency order:

```go
func main() {
    // 1. Initialize database
    database.Init(dbPath, useWAL)

    // 2. Inject stores into service layer
    service.SetLedgerStore(database.NewLedgerStore())
    service.SetKeyStore(database.NewKeyStore())

    // 3. Initialize services with configuration
    service.InitKeys(bootstrapKey)
    service.InitCORS(allowedOrigins)

    // 4. Build and serve HTTP router
    r := mux.NewRouter()
    api.BuildRouter(r.PathPrefix("/api/v1").Subrouter())
    http.ListenAndServe(port, r)
}
```

## Middleware Pattern

Write composable middleware that wraps handlers.

```go
func withAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := extractBearerToken(r)
        if ok, err := service.VerifyAPIKey(token); !ok || err != nil {
            writeError(w, http.StatusUnauthorized, "Unauthorized")
            return
        }
        next(w, r)
    }
}

// Apply to routes
r.HandleFunc("/protected", withAuth(handleProtected)).Methods("POST")
```

## Database Migrations

Use schema versioning with sequential migrations.

```go
type migration struct {
    version int
    sql     string
}

var migrations = []migration{
    {version: 1, sql: "CREATE TABLE..."},
    {version: 2, sql: "ALTER TABLE..."},
}

func runMigrations() {
    current := getSchemaVersion() // PRAGMA user_version
    for _, m := range migrations {
        if m.version > current {
            db.Exec(m.sql)
            setSchemaVersion(m.version)
        }
    }
}
```

## CLI Structure

Import and use the `git.sr.ht/~jakintosh/command-go` package to implement command line argument parsing.

Organize commands hierarchically with global options.

```go
var root = &cmd.Command{
    Subcommands: []*cmd.Command{
        apiCmd,     // HTTP API calls
        envCmd,     // Environment management
        serveCmd,   // Start server
    },
    Options: []cmd.Option{
        {Long: "url", Type: cmd.OptionTypeParameter},
        {Long: "env", Type: cmd.OptionTypeParameter},
    },
}
```

**Nested structure example**:
```
api/
├── ledger/
│   ├── snapshot
│   └── tx/
│       ├── list
│       └── create
└── settings/
    ├── keys/
    │   ├── create
    │   └── delete
    └── cors/
        ├── get
        └── set
```

## Configuration Resolution

Apply priority order: **CLI params > Environment > Config file > Defaults**

```go
func resolveOption(cliParam, envVar, defaultValue string) string {
    if cliParam != "" {
        return cliParam
    }
    if env := os.Getenv(envVar); env != "" {
        return env
    }
    if configValue := loadFromConfig(); configValue != "" {
        return configValue
    }
    return defaultValue
}
```

## Error Handling

Define domain-specific errors as sentinels.

```go
var (
    ErrInvalidAllocation = errors.New("invalid allocation percentages")
    ErrNoLedgerStore     = errors.New("ledger store not configured")
)

// Custom error type for wrapped errors
type DatabaseError struct{ Err error }
func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }
```

**Check errors with type assertions**:
```go
if err != nil {
    var dbErr service.DatabaseError
    if errors.As(err, &dbErr) {
        // Handle database error
    } else if errors.Is(err, service.ErrNoStore) {
        // Handle missing dependency
    }
}
```

## HTTP Status Codes

Map errors to appropriate status codes:

- `200 OK` - Successful GET
- `201 Created` - Successful POST
- `204 No Content` - Successful PUT/DELETE
- `400 Bad Request` - Malformed input
- `401 Unauthorized` - Invalid/missing auth
- `403 Forbidden` - CORS or policy violation
- `500 Internal Server Error` - Database/logic errors
- `503 Service Unavailable` - Database unreachable

## Naming Conventions

- **Interfaces**: `XxxStore`, `XxxService`
- **Implementations**: `DBXxxStore`
- **Constructors**: `NewXxxStore()`
- **Handlers**: `handleGetXxx`, `handlePostXxx`
- **Router builders**: `buildXxxRouter()`
- **Middleware**: `withXxx()`
- **Setters**: `SetXxxStore()`

## Testing Patterns

Provide test utilities for setup and assertions.

```go
func setupRouter() *mux.Router {
    router := mux.NewRouter()
    api.BuildRouter(router)
    return router
}

func makeAuthHeader(token string) map[string]string {
    return map[string]string{"Authorization": "Bearer " + token}
}

// Generic request helper
func get(router *mux.Router, url string, response any, headers map[string]string) httpResult {
    // Execute request, unmarshal response
}
```

## Security Practices

**API Keys**:
- Format: `{id}.{secret}` (e.g., `default.a1b2c3d4...`)
- Storage: SHA256 hash with random salt
- Verification: Constant-time comparison

**CORS**:
- Whitelist allowed origins in database
- Validate `Origin` header on all requests
- Return `403 Forbidden` for disallowed origins

**Credentials**:
- Load sensitive values from files in credentials directory
- Default to `/etc/project` for systemd integration
- Support override via `--credentials-directory` flag

## Data Patterns

**Timestamps**: Store as Unix epoch integers, convert to `time.Time` at service layer

**Pagination**: Standard `limit` (default 100) and `offset` (default 0) parameters

**Upserts**: Use `INSERT ... ON CONFLICT(id) DO UPDATE SET ...` for idempotency

**Safe aggregation**: Use `COALESCE(SUM(column), 0)` to handle NULL results

**Parameterized queries**: Always use placeholders (`?1`, `?2`) to prevent SQL injection

## SQLite Configuration

```go
db.SetMaxOpenConns(1)                    // Serial writes
db.Exec("PRAGMA journal_mode=WAL")       // Write-Ahead Logging
db.Exec("PRAGMA busy_timeout=5000")      // 5s wait on locks
db.Exec("PRAGMA foreign_keys=ON")        // Referential integrity
```

## Health Check Pattern

Implement transactional health probe for database connectivity.

```go
func HealthCheck() error {
    // Create temp table
    // Insert row
    // Read row
    // Drop table
    // Return error if any step fails
}
```

---

**Summary**: Keep layers separate, use interfaces for DI, compose middleware, validate early, handle errors explicitly, and maintain consistent naming. This structure scales from small utilities to production services.
