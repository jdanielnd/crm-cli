# crm-cli Go Implementation Plan

## Phase 1 — Project Scaffold & Person CRUD (manually testable end-to-end)

Set up the Go project from scratch and implement the first entity (People) with full CRUD, output formatting, and tests. By the end of this phase you can `go run ./cmd/crm person add "Jane Smith"` and see results.

**Steps:**
1. `go mod init`, install dependencies (cobra, modernc.org/sqlite, testify, etc.)
2. Create `cmd/crm/main.go` — entry point, version injection, signal handling
3. Create `internal/model/types.go` — Person struct, constants (interaction types, deal stages, priorities, relationship types)
4. Create `internal/model/errors.go` — sentinel errors, ExitError, ExitCode mapper
5. Create `internal/db/db.go` — Open(), pragmas, migration runner with `go:embed`
6. Create `internal/db/migrations/001_initial.sql` — full schema (all tables, FTS5, triggers)
7. Create `internal/db/repo/person.go` — Create, FindByID, FindAll, Update, Archive, Search (FTS5)
8. Create `internal/format/` — format resolver, JSON, table, CSV formatters
9. Create `internal/cli/root.go` — root Cobra command, global persistent flags (--format, --quiet, --verbose, --db, --no-color)
10. Create `internal/cli/person.go` — person add, list, show, edit, delete subcommands
11. Create `.gitignore`, `.editorconfig`
12. Unit tests for db, repo/person, format, model/errors
13. Integration test — build binary, run person commands, assert output

## Phase 2 — Organization CRUD

Add organizations as the second entity, with person-org linking.

1. `internal/db/repo/org.go` — Create, FindByID, FindAll, Update, Archive, Search
2. `internal/cli/org.go` — org add, list, show (--with people), edit, delete
3. Unit + integration tests

## Phase 3 — Interaction Logging

Log calls, emails, meetings, notes, messages linked to people.

1. `internal/db/repo/interaction.go` — Create (with person_ids junction), FindAll, Search
2. `internal/cli/log.go` — `crm log call|email|meeting|note|message <person_ids> --subject --content --direction --at`
3. Unit + integration tests

## Phase 4 — Tags & Custom Fields

Polymorphic tagging system and key/value custom fields.

1. `internal/db/repo/tag.go` — Apply, Remove, GetForEntity, FindAll
2. `internal/db/repo/custom_field.go` — Set, Get, Delete
3. `internal/cli/tag.go` — tag list, apply, remove, show, delete
4. Unit + integration tests

## Phase 5 — Relationships

Person-to-person links (colleague, friend, manager, mentor, referred-by).

1. `internal/db/repo/relationship.go` — Create, FindForPerson, Delete
2. `internal/cli/relate.go` — person relate, person relationships, person unrelate
3. Unit + integration tests

## Phase 6 — Deals & Pipeline

Deal tracking with stage progression and pipeline summary.

1. `internal/db/repo/deal.go` — Create, FindByID, FindAll (stage filter), Update, Archive, Pipeline, Search
2. `internal/cli/deal.go` — deal add, list, show, edit, delete, pipeline
3. Unit + integration tests

## Phase 7 — Tasks & Follow-ups

Task management with due dates, priorities, and completion.

1. `internal/db/repo/task.go` — Create, FindAll (overdue/pending/done), Complete, Update, Archive
2. `internal/cli/task.go` — task add, list, show, edit, done, delete
3. Unit + integration tests

## Phase 8 — Search, Context & Status

Cross-entity search, person briefing, and dashboard.

1. `internal/cli/search.go` — `crm search <query>` across people, orgs, interactions, deals
2. `internal/cli/context.go` — `crm context <id>` full person briefing
3. `internal/cli/status.go` — `crm status` dashboard summary
4. Unit + integration tests

## Phase 9 — MCP Server

Built-in MCP server for AI agent integration.

1. `internal/mcp/server.go` — MCP server with all 18 tools
2. `internal/cli/mcp.go` — `crm mcp serve` command
3. Integration tests with MCP inspector

## Phase 10 — CI/CD & Release

GitHub Actions, GoReleaser, Homebrew tap.

1. `.github/workflows/ci.yml` — lint, test, build (matrix: ubuntu, macos)
2. `.goreleaser.yml` — cross-platform builds, Homebrew tap
3. `.github/workflows/release.yml` — tag-triggered release
4. `.golangci.yml` — linter config
