# CLAUDE.md — crm-cli

## Project Overview

`crm` is a local-first personal CRM for the terminal. Go, SQLite (`modernc.org/sqlite`), Cobra CLI, built-in MCP server. Distributed as a single static binary for macOS, Linux, and Windows. See `README.md` for full product spec and `overview.md` for the technology-agnostic product vision.

## Quick Reference

- **Language:** Go 1.23+
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **CLI framework:** `cobra`
- **MCP:** `mcp-go` (stdio transport)
- **Testing:** `testing` + `testify`
- **Linting:** `golangci-lint`
- **Build:** `go build`
- **Releases:** GoReleaser + conventional commits

## Commands

```bash
go run ./cmd/crm              # Run CLI in development
go build -o crm ./cmd/crm     # Build binary
go test ./...                  # Run all tests
go test ./internal/db/repo/... # Unit tests for repos only
go test ./internal/cli/...     # CLI integration tests
go test -race ./...            # Tests with race detector
go test -v -run TestPerson ./... # Run specific test
golangci-lint run              # Lint
go vet ./...                   # Static analysis
```

## Project Structure

```
cmd/
  crm/
    main.go                    # Entry point, signal handling
internal/
  cli/
    root.go                    # Root cobra command, global persistent flags
    person.go                  # Person subcommands (add, list, show, edit, delete)
    org.go                     # Organization subcommands
    log.go                     # Interaction logging (call, email, meeting, note, message)
    tag.go                     # Tag management
    relate.go                  # Person-to-person relationships
    deal.go                    # Deals & pipeline
    task.go                    # Tasks & follow-ups
    search.go                  # Cross-entity full-text search
    context.go                 # Person context briefing
    status.go                  # Dashboard summary
    mcp.go                     # MCP server launcher
  db/
    db.go                      # Connection, pragmas, migration runner
    migrations/                # Embedded SQL files (go:embed)
      001_initial.sql
    repo/
      person.go                # Person repository
      org.go                   # Organization repository
      interaction.go           # Interaction repository
      tag.go                   # Tag repository (polymorphic)
      deal.go                  # Deal repository
      task.go                  # Task repository
      custom_field.go          # Custom field repository
      relationship.go          # Person-to-person relationship repository
  mcp/
    server.go                  # MCP server setup + all tool definitions
  format/
    format.go                  # Format resolver (table/json/csv/tsv)
    table.go                   # Table formatter
    json.go                    # JSON formatter
    csv.go                     # CSV/TSV formatter
  model/
    types.go                   # Domain types, constants, enums
    errors.go                  # Error types with exit codes
```

## Architecture Rules

### Layers

1. **CLI commands** (`internal/cli/`) — parse args via Cobra, call repositories, format output. No direct SQL.
2. **Repositories** (`internal/db/repo/`) — all database queries. Accept and return typed structs. Use parameterized queries.
3. **Formatters** (`internal/format/`) — transform data to table/json/csv. No business logic.
4. **Models** (`internal/model/`) — domain types, constants, error types. No imports from other internal packages.
5. **MCP server** (`internal/mcp/`) — thin wrappers around the same repositories the CLI uses. No duplicate logic.

### Key Patterns

- **Single DB connection** — `internal/db/db.go` opens the connection, sets pragmas, runs migrations. Pass `*sql.DB` down to repositories via dependency injection (not a global).
- **Parameterized queries** — always use `?` placeholders. Never interpolate user input into SQL strings.
- **Transactions** — wrap multi-step mutations in `db.BeginTx()` / `tx.Commit()`.
- **Soft deletes** — set `archived = 1`, never `DELETE FROM`. Filter `WHERE archived = 0` by default in all queries.
- **Integer IDs + UUIDs** — integer `id` for CLI/human use, `uuid` column for external/API references. Generate UUIDs with `google/uuid`.
- **Exit codes** — 0=success, 1=error, 2=usage error, 3=not found, 4=conflict, 10=db error. Use `os.Exit()` only in `main.go`; commands return errors that bubble up.
- **stdout/stderr discipline** — data output (tables, JSON, CSV) goes to `os.Stdout` only. All human messages (success confirmations, progress, warnings, errors) go to `os.Stderr`. Critical for piping.
- **Output format** — respect `--format` flag everywhere. Default to table for TTY, JSON for pipes. Check with `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))`.
- **Error handling** — return `error` from all functions. Define sentinel errors and typed errors in `model/errors.go`. Cobra's `RunE` propagates errors to the root command which formats them.
- **Signal handling** — handle `SIGINT`/`SIGTERM` in `main.go` via `signal.NotifyContext`. Close DB connection in a deferred cleanup. Exit with `128 + signal_number`.
- **Context propagation** — pass `context.Context` through CLI commands to repositories for cancellation support.

### Error Types

```go
// Sentinel errors for classification
var (
    ErrNotFound   = errors.New("not found")
    ErrValidation = errors.New("validation error")
    ErrConflict   = errors.New("conflict")
    ErrDatabase   = errors.New("database error")
)

// ExitError wraps a sentinel with a user-facing message
type ExitError struct {
    Message string
    Err     error
}

func (e *ExitError) Error() string { return e.Message }
func (e *ExitError) Unwrap() error { return e.Err }

// Map sentinel errors to exit codes at the top level
func ExitCode(err error) int {
    switch {
    case errors.Is(err, ErrValidation): return 2
    case errors.Is(err, ErrNotFound):   return 3
    case errors.Is(err, ErrConflict):   return 4
    case errors.Is(err, ErrDatabase):   return 10
    default:                            return 1
    }
}
```

Use `fmt.Errorf("person %d: %w", id, ErrNotFound)` to wrap sentinels with context. At the top level, use `errors.Is()` to classify and `ExitCode()` to map. Print human-readable messages to stderr prefixed with `crm: error:`. Show wrapped cause only with `--verbose`.

## Go Conventions

### Module Path

```
module github.com/jdanielnd/crm-cli
```

### Naming

- Files: `snake_case.go` (e.g., `person.go`, `custom_field.go`)
- Types/interfaces: `PascalCase` (e.g., `Person`, `PersonRepo`, `CreatePersonInput`)
- Functions: `PascalCase` for exported, `camelCase` for unexported
- Constants: `PascalCase` for exported, `camelCase` for unexported
- Database columns: `snake_case`
- CLI flags: `--kebab-case`

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `golangci-lint` with these linters: `errcheck`, `govet`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, `gocritic`, `revive`
- **No `init()` functions** — explicit initialization in `main.go` or constructors
- **No package-level mutable state** — pass dependencies explicitly
- **Interfaces at the consumer** — define small interfaces where they're used, not where they're implemented
- **Errors are values** — return `error`, don't panic. Reserve `panic` for truly unrecoverable programmer errors
- **Table-driven tests** — use subtests with `t.Run()` for related test cases
- **`internal/` package** — all application code goes in `internal/` to prevent external imports

### Struct Patterns

```go
// Domain type — maps to database row
type Person struct {
    ID        int64   `json:"id"`
    UUID      string  `json:"uuid"`
    FirstName string  `json:"first_name"`
    LastName  *string `json:"last_name"`
    Email     *string `json:"email"`
    CreatedAt string  `json:"created_at"`
    UpdatedAt string  `json:"updated_at"`
    Archived  bool    `json:"-"`
}

// Input type — for creating
type CreatePersonInput struct {
    FirstName string
    LastName  *string
    Email     *string
    Phone     *string
    // ...
}

// Input type — for updating (all pointer fields = optional/patch semantics)
type UpdatePersonInput struct {
    FirstName *string
    LastName  *string
    Email     *string
    // ...
}
```

Use `*string` (pointer) for nullable database columns and optional update fields. Use `string` for required fields.

## Database Conventions

### Connection Setup

Set these pragmas immediately after opening:

```sql
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;
PRAGMA cache_size = -64000;
PRAGMA temp_store = MEMORY;
```

### Pure-Go SQLite

Use `modernc.org/sqlite` — pure Go, no CGO required. This is critical for clean cross-compilation. Import as:

```go
import _ "modernc.org/sqlite"
// then use database/sql
db, err := sql.Open("sqlite", dbPath)
```

### Migrations

- Embedded via `//go:embed migrations/*.sql` in `internal/db/db.go`
- Sequential SQL files: `001_initial.sql`, `002_add_summary.sql`, etc.
- Each file has `-- up` and `-- down` sections
- Track current version via `PRAGMA user_version`
- Run inside a transaction for atomicity
- Auto-backup database file before running migrations
- **Never modify existing migrations** — always create a new one

### Schema

- **FTS5** indexes on people, organizations, interactions, deals — keep in sync via triggers
- **Timestamps** — store as ISO 8601 strings (`TEXT`). Use `datetime('now')` for defaults
- **Booleans** — `INTEGER` (0/1)

### Query Patterns

```go
// Single row
var p Person
err := db.QueryRowContext(ctx, "SELECT ... FROM people WHERE id = ?", id).
    Scan(&p.ID, &p.UUID, &p.FirstName, &p.LastName, &p.Email, &p.CreatedAt, &p.UpdatedAt, &p.Archived)
if errors.Is(err, sql.ErrNoRows) {
    return nil, fmt.Errorf("person %d: %w", id, ErrNotFound)
}

// Multiple rows
rows, err := db.QueryContext(ctx, "SELECT ... FROM people WHERE archived = 0")
if err != nil {
    return nil, fmt.Errorf("list people: %w", err)
}
defer rows.Close()
for rows.Next() {
    var p Person
    if err := rows.Scan(&p.ID, &p.FirstName, ...); err != nil {
        return nil, fmt.Errorf("scan person: %w", err)
    }
    results = append(results, p)
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterate people: %w", err)
}

// Exec
result, err := db.ExecContext(ctx, "INSERT INTO people (...) VALUES (...)", ...)
if err != nil {
    return 0, fmt.Errorf("insert person: %w", err)
}
id, _ := result.LastInsertId()
```

Start with `database/sql` and manual `Scan()` calls. If boilerplate becomes excessive, consider `jmoiron/sqlx` for struct scanning via `db:"column_name"` tags.

## Testing

- **Unit tests** (`*_test.go` next to source) — test repositories and formatters with in-memory SQLite (`:memory:`)
- **Integration tests** (`internal/cli/*_test.go`) — build the binary, exec it with `os/exec`, assert on stdout/stderr/exit code. Use `t.TempDir()` for isolated DB paths.
- **Table-driven tests** — standard Go pattern with `t.Run()` subtests
- **Each test gets a fresh DB** — open `:memory:`, run migrations, seed, test. Instant setup since SQLite is embedded.
- **No mocking the database** — use real SQLite in-memory
- **Test helpers** — create a `testutil` package or test helpers in `*_test.go` files for common setup (open DB, run migrations, seed data)
- **`-race` flag** — always run tests with race detector in CI
- **Test binary** — for integration tests, build once with `go build` in `TestMain`, reuse across subtests

```go
func TestPersonRepo_Create(t *testing.T) {
    db := testutil.NewTestDB(t)
    repo := repo.NewPersonRepo(db)

    tests := []struct {
        name  string
        input CreatePersonInput
        want  string
    }{
        {"basic", CreatePersonInput{FirstName: "Jane"}, "Jane"},
        {"with email", CreatePersonInput{FirstName: "Bob", Email: ptr("bob@test.com")}, "Bob"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p, err := repo.Create(context.Background(), tt.input)
            require.NoError(t, err)
            assert.Equal(t, tt.want, p.FirstName)
        })
    }
}
```

## MCP Server

- Uses `mcp-go` with stdio transport
- **Never write to stdout** in MCP mode — stdout is the JSON-RPC transport. All logging goes to `stderr`.
- Each MCP tool is a thin wrapper calling repository methods
- Return errors as MCP content with `isError: true` — don't return Go errors from tool handlers
- Handle graceful shutdown via context cancellation
- Tool input is JSON-decoded into Go structs

## Date Handling

- Store all dates as ISO 8601 UTC strings in SQLite
- Parse natural language dates with a Go library (e.g., `olebedev/when` or `tj/go-naturaldate`)
- Fall back to standard time.Parse with common formats
- Display in local time at output
- For tests, inject a reference `time.Time` for deterministic behavior

## Data Directory

Default database path resolution:

1. `--db <path>` flag (highest priority)
2. `CRM_DB` environment variable
3. `~/.crm/crm.db` (default)

On macOS, suggest iCloud path for backup:
```
~/Library/Mobile Documents/com~apple~CloudDocs/crm/crm.db
```

Auto-create the directory and database file on first use.

## Build & Release

### Local Build

```bash
go build -o crm ./cmd/crm
```

### GoReleaser

Use GoReleaser for cross-platform builds and GitHub Releases:

```yaml
# .goreleaser.yml
builds:
  - main: ./cmd/crm
    binary: crm
    env: [CGO_ENABLED=0]
    goos: [darwin, linux, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.ShortCommit}}
```

### Version Injection

Inject version at build time via `-ldflags`:

```go
// cmd/crm/main.go
var (
    version = "dev"
    commit  = "none"
)
```

## Git Conventions

### Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(person): add merge command
fix(db): handle WAL lock contention
docs: update README with new examples
refactor(repo): extract base repository
test(deal): add pipeline stage transition tests
feat(cli)!: rename log command to interact    # breaking change
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`

### Branches

```
feat/add-person-merge
fix/wal-lock-contention
docs/contributing-guide
```

### PR Workflow

- All work on feature branches off `main`
- `main` is always releasable
- Squash merge into `main`
- Delete branches after merge
- CI must pass before merge

## CI/CD (GitHub Actions)

### `ci.yml` — every push and PR

1. **Lint** — `golangci-lint run`
2. **Test** — `go test -race ./...` on Go 1.23 (matrix: ubuntu, macos)
3. **Build** — `go build ./cmd/crm`

### `release.yml` — on tag push

Runs GoReleaser to build binaries for all platforms, create GitHub Release with artifacts, and update Homebrew tap.

## Repository Files

- `.gitignore` — binary outputs, `*.db`, `.env`, `.DS_Store`, `dist/`
- `.golangci.yml` — linter configuration
- `.goreleaser.yml` — release configuration
- `.editorconfig` — tabs for Go, UTF-8, trim trailing whitespace
- `LICENSE` — MIT

## Common Tasks

### Adding a new entity

1. Define types in `internal/model/types.go` (row struct, create input, update input)
2. Create migration in `internal/db/migrations/`
3. Create repository in `internal/db/repo/`
4. Create CLI commands in `internal/cli/`
5. Register commands in `internal/cli/root.go`
6. Add MCP tools in `internal/mcp/server.go`
7. Add tests for repository (unit) + CLI command (integration)

### Adding a new CLI command to an existing entity

1. Add the Cobra subcommand in the entity's CLI file
2. Call the appropriate repository method
3. Format output through the formatter
4. Add unit test for the repository method
5. Add integration test for the CLI command

### Adding a new MCP tool

1. Define input struct in `internal/mcp/server.go` (or reuse from `model/`)
2. Register tool with description and handler
3. Wire to existing repository method
4. Test with MCP inspector
