# Code Review: crm-cli

**Date:** 2026-03-09
**Overall Grade:** A- (Very Good)

---

## Issues Found

### HIGH Priority

#### H1. `status.go` bypasses repository layer and swallows errors

**Location:** `internal/cli/status.go:28-57, 90, 102`

The status command executes raw SQL directly against `*sql.DB`, violating the architecture rule in CLAUDE.md that CLI commands should "call repositories, format output. No direct SQL." Additionally, 8 instances of `_ =` silently discard database errors — if the DB is corrupt or locked, the dashboard shows zeros instead of reporting the problem.

```go
// Current: raw SQL + swallowed error
var personCount int
_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM people WHERE archived = 0").Scan(&personCount)

// Also swallowed:
stages, _ := dr.Pipeline(ctx)
tasks, _ := tr.FindAll(ctx, model.TaskFilters{Overdue: true, Limit: 5})
```

**Fix:** Add `Count(ctx)` methods to each relevant repo and use them. Handle all errors properly.

---

#### H2. ID parsing duplicated 20+ times across CLI files

**Location:** `internal/cli/person.go`, `org.go`, `deal.go`, `task.go`, `log.go`, `tag.go`, `relate.go`, `context.go`

Every command that takes an ID argument repeats the same 4-line pattern:

```go
id, err := strconv.ParseInt(args[0], 10, 64)
if err != nil {
    return model.NewExitError(model.ErrValidation, "invalid person ID: %s", args[0])
}
```

A multi-ID parsing loop is also duplicated in `log.go`, `tag.go`, and `relate.go`:

```go
var personIDs []int64
for _, arg := range args {
    id, err := strconv.ParseInt(arg, 10, 64)
    if err != nil {
        return model.NewExitError(model.ErrValidation, "invalid person ID: %s", arg)
    }
    personIDs = append(personIDs, id)
}
```

There is already a `parseID` function in `context.go:234` but it is unexported and not reused.

**Fix:** Extract shared `parseID(s string, entity string) (int64, error)` and `parseIDs(args []string, entity string) ([]int64, error)` helpers in a common location (e.g., `internal/cli/helpers.go`).

---

### MEDIUM Priority

#### M1. MCP test coverage is very thin

**Location:** `internal/mcp/server_test.go`

Only 3 tests exist for 30+ MCP tools. There are no error path tests (not found, conflict, validation), no tool invocation variety beyond person create, and no tests for array parameter handling.

**Fix:** Add table-driven tests covering happy path and error paths for each tool category (person, org, interaction, deal, task, tag, relationship, search, context, stats).

---

#### M2. CSV headers use field names instead of display names

**Location:** `internal/format/format.go:127`

CSV output uses `col.Field` (e.g., `first_name`) for headers while table output uses `col.Header` (e.g., `First Name`). This inconsistency confuses users switching between formats.

```go
// CSV (line 127) — uses raw field name
headers[i] = col.Field

// Table (line 100) — uses display name
header[i] = col.Header
```

**Fix:** Change CSV header to use `col.Header` to match table output.

---

#### M3. `outputJSON` TTY detection hardcodes `os.Stdout`

**Location:** `internal/format/format.go:82`

The function accepts `io.Writer` but checks `os.Stdout` for TTY detection regardless of what writer was passed. In tests or when writing to a buffer, this produces incorrect behavior.

```go
func outputJSON(w io.Writer, data []map[string]any) error {
    enc := json.NewEncoder(w)
    if term.IsTerminal(int(os.Stdout.Fd())) {  // should check w, not os.Stdout
        enc.SetIndent("", "  ")
    }
    return enc.Encode(data)
}
```

**Fix:** Pass a `isTTY bool` parameter to `outputJSON` (and `Output`), resolved once at the call site.

---

### LOW Priority

#### L1. Validator functions use linear search

**Location:** `internal/model/types.go:236-308`

Six `Valid*()` functions all perform identical O(n) linear scans over small slices. While performance is not a concern for these small lists, the code is repetitive (6 identical function bodies).

```go
func ValidEntityType(t string) bool {
    for _, et := range EntityTypes {
        if et == t { return true }
    }
    return false
}
// Same pattern repeated 5 more times
```

**Fix:** Extract a generic `contains(slice []string, value string) bool` helper or use map-based lookup sets.

---

#### L2. Package-level mutable flag variables

**Location:** `internal/cli/root.go:14-19`

```go
var (
    flagFormat  string
    flagQuiet   bool
    flagVerbose bool
    flagDB      string
    flagNoColor bool
)
```

These package-level variables make the CLI package untestable in-process (can't run two commands concurrently) and violate the CLAUDE.md guideline "No package-level mutable state." The integration tests work around this by shelling out to the binary.

**Fix:** Move flag variables into a struct passed through Cobra's context or attached to the root command. This is a larger refactor and may not be worth the disruption right now.

---

#### L3. FTS5 search pattern duplicated across 4 repos

**Location:** `internal/db/repo/person.go:249-279`, `org.go:170-199`, `interaction.go:178-217`, `deal.go:187-213`

All four searchable repos have nearly identical `Search()` methods: same FTS escaping, same limit defaulting, same row scanning loop.

```go
// Identical in all 4 repos:
if limit <= 0 { limit = 20 }
ftsQuery := `"` + strings.ReplaceAll(query, `"`, `""`) + `"` + "*"
```

**Fix:** Extract a shared `ftsSearch` helper that handles escaping, limit defaults, and query construction, leaving only the entity-specific column scanning to each repo.

---

#### L4. Nil-to-string conversion duplicated in formatter

**Location:** `internal/format/format.go:107-112` and `136-141`

The same nil-check + `fmt.Sprintf` pattern appears in both `outputTable` and `outputCSV`:

```go
v := row[col.Field]
if v == nil {
    vals[i] = ""
} else {
    vals[i] = fmt.Sprintf("%v", v)
}
```

**Fix:** Extract a `formatValue(v any) string` helper.

---

#### L5. `Resolve()` and `outputJSON()` both call `term.IsTerminal()`

**Location:** `internal/format/format.go:38` and `format.go:82`

`term.IsTerminal()` is called twice per output cycle — once in `Resolve()` and again in `outputJSON()`. While cheap, resolving once would be cleaner.

**Fix:** Resolve TTY status once and thread it through as a parameter.

---

## Summary Table

| ID | Priority | Category | Description |
|----|----------|----------|-------------|
| H1 | High | Architecture | `status.go` bypasses repo layer + swallows errors |
| H2 | High | DRY | ID parsing duplicated 20+ times |
| M1 | Medium | Testing | MCP test coverage is very thin (3/30 tools) |
| M2 | Medium | Consistency | CSV headers use field names, not display names |
| M3 | Medium | Correctness | `outputJSON` TTY check hardcodes `os.Stdout` |
| L1 | Low | DRY | Validator functions repeat linear search |
| L2 | Low | Architecture | Package-level mutable flag variables |
| L3 | Low | DRY | FTS5 search duplicated across 4 repos |
| L4 | Low | DRY | Nil-to-string duplicated in formatter |
| L5 | Low | Efficiency | `term.IsTerminal()` called redundantly |
