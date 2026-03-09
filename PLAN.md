# Implementation Plan: Code Quality Fixes

This plan addresses all issues identified in `CODE_REVIEW.md`, ordered by priority and dependency.

---

## Phase 1: Foundation Helpers (no breaking changes)

These create shared utilities that later phases depend on.

### Step 1.1 — Extract `formatValue` helper in formatter (L4)

**File:** `internal/format/format.go`

- Add `func formatValue(v any) string` that handles nil → `""` and non-nil → `fmt.Sprintf("%v", v)`
- Replace the duplicated nil-check blocks in `outputTable` (lines 107-112) and `outputCSV` (lines 136-141) with calls to `formatValue`
- Run `go test ./internal/format/...` to verify

### Step 1.2 — Extract `contains` helper for validators (L1)

**File:** `internal/model/types.go`

- Add unexported `func contains(slice []string, value string) bool`
- Rewrite all 6 `Valid*()` functions to use it: `ValidEntityType`, `ValidInteractionType`, `ValidInteractionDirection`, `ValidDealStage`, `ValidPriority`, `ValidRelationshipType`
- Run `go test ./internal/model/...` to verify

### Step 1.3 — Extract `parseID` and `parseIDs` CLI helpers (H2)

**File:** Create `internal/cli/helpers.go`

- Move the existing unexported `parseID` from `context.go:234` into `helpers.go`
- Generalize signature: `func parseID(s string, entity string) (int64, error)` — returns `model.NewExitError(model.ErrValidation, "invalid %s ID: %s", entity, s)`
- Add `func parseIDs(args []string, entity string) ([]int64, error)` for multi-ID parsing
- Update all call sites in: `person.go`, `org.go`, `deal.go`, `task.go`, `log.go`, `tag.go`, `relate.go`, `context.go`
- Remove the old `parseID` from `context.go`
- Run `go test ./internal/cli/...` to verify

---

## Phase 2: Formatter Fixes

### Step 2.1 — Fix CSV headers to use display names (M2)

**File:** `internal/format/format.go`

- Change line 127 from `headers[i] = col.Field` to `headers[i] = col.Header`
- Update `format_test.go` to assert CSV headers use display names
- Run `go test ./internal/format/...` to verify

### Step 2.2 — Thread TTY detection through formatter (M3, L5)

**File:** `internal/format/format.go`

- Add `isTTY bool` parameter to `Output()` function signature: `func Output(w io.Writer, format Format, data []map[string]any, columns []ColumnDef, quiet bool, isTTY bool) error`
- Pass `isTTY` to `outputJSON` instead of having it check `os.Stdout.Fd()` internally
- Remove the `term.IsTerminal` call from inside `outputJSON`
- Update `Resolve()` to also accept a `isTTY bool` parameter (or keep it checking stdout since it's only called from CLI context)
- Update all call sites in `internal/cli/*.go` to pass `term.IsTerminal(int(os.Stdout.Fd()))` (can be computed once per command)
- Update `format_test.go`
- Run `go test ./...` to verify

---

## Phase 3: Status Command Refactor (H1)

### Step 3.1 — Add count/summary methods to repositories

**Files:**
- `internal/db/repo/person.go` — add `func (r *PersonRepo) Count(ctx context.Context) (int, error)`
- `internal/db/repo/org.go` — add `func (r *OrgRepo) Count(ctx context.Context) (int, error)`
- `internal/db/repo/deal.go` — add `func (r *DealRepo) OpenSummary(ctx context.Context) (count int, totalValue float64, err error)`
- `internal/db/repo/task.go` — add `func (r *TaskRepo) OverdueCount(ctx context.Context) (int, error)` and `func (r *TaskRepo) OpenCount(ctx context.Context) (int, error)`
- `internal/db/repo/interaction.go` — add `func (r *InteractionRepo) CountSince(ctx context.Context, since string) (int, error)`

Add tests for each new method in the corresponding `*_test.go` files.

### Step 3.2 — Rewrite `status.go` to use repositories

**File:** `internal/cli/status.go`

- Replace all raw SQL queries with calls to the new repo methods from Step 3.1
- Handle all errors properly (no more `_ =` discards)
- Keep the same output format and behavior
- Run `go test ./internal/cli/...` to verify

---

## Phase 4: FTS Search Deduplication (L3)

### Step 4.1 — Extract shared FTS helper

**File:** Create `internal/db/repo/fts.go`

- Add `func escapeFTS(query string) string` — handles the `"` escaping and `*` suffix
- Add `func defaultLimit(limit int) int` — returns 20 if limit <= 0
- Update `Search()` in `person.go`, `org.go`, `interaction.go`, `deal.go` to use these helpers
- Run `go test ./internal/db/repo/...` to verify

---

## Phase 5: MCP Test Coverage (M1)

### Step 5.1 — Expand MCP server tests

**File:** `internal/mcp/server_test.go`

Add table-driven tests for:
- **Person tools:** search, get, create, update, delete — including not-found and validation errors
- **Organization tools:** search, get
- **Interaction tools:** log (with array person_ids), list
- **Deal tools:** create (with validation), update
- **Task tools:** create, list, complete
- **Tag tools:** apply (with invalid entity type)
- **Search tool:** cross-entity search
- **Context tool:** full briefing
- **Stats tool:** dashboard summary

Each test should cover:
1. Happy path (valid input → expected output)
2. Error path: missing required arguments
3. Error path: entity not found
4. Error path: validation errors (invalid enum values, etc.)

---

## Phase 6: Package-Level State (L2) — Optional / Deferred

### Step 6.1 — Move flags into a config struct

**Files:** `internal/cli/root.go` and all CLI files

- Define a `type cliConfig struct { Format, DB string; Quiet, Verbose, NoColor bool }`
- Store in Cobra's command context or as a field on a wrapper struct
- Update all commands to read from the struct instead of package vars
- This is a large, cross-cutting refactor — defer unless there's a concrete need for in-process testing

---

## Execution Order

```
Phase 1 (helpers)     → no dependencies, safe to do first
Phase 2 (formatter)   → depends on 1.1 (formatValue)
Phase 3 (status.go)   → independent of phases 1-2
Phase 4 (FTS dedup)   → independent
Phase 5 (MCP tests)   → independent
Phase 6 (flag struct) → deferred
```

Phases 1-4 can be implemented as separate commits. Phase 5 (tests) can be done in parallel. Phase 6 is optional.

---

## Verification

After all changes:

```bash
go test ./...                  # All tests pass
go test -race ./...            # No race conditions
golangci-lint run              # No lint issues
go vet ./...                   # No static analysis issues
go build ./cmd/crm             # Binary builds
```
