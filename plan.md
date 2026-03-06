# crm-cli v1 Implementation Plan

## Goal

Ship a usable CRM CLI that covers the core workflow: manage contacts and organizations, log interactions, tag entities, search across everything, get AI-ready context on a person, and expose it all via MCP. This is Phases 1-3 from the README, plus tags and tasks from Phase 2/4 — everything needed for the tool to be genuinely useful day-to-day.

## What's in v1

- People CRUD (add, list, show, edit, delete)
- Organization CRUD (add, list, show, edit, delete)
- Interaction logging (call, email, meeting, note)
- Tags (apply, remove, list — polymorphic on any entity)
- Tasks (add, list, done — tied to people)
- Deals (add, list, edit, pipeline — tied to people/orgs)
- Custom fields (key/value on any entity)
- Person-to-person relationships
- Full-text search across all entities
- `crm context` command (the AI briefing)
- `crm status` dashboard
- MCP server with full tool suite
- Table, JSON, CSV, TSV output formats
- `--quiet` mode for piping

## What's NOT in v1

- Import/export (CSV, vCard)
- Duplicate detection / merge
- Shell completions
- fzf integration
- Bulk operations
- Man page
- Auto-backups

---

## Step 0: Project Scaffolding

Set up the project infrastructure before writing any feature code.

### 0.1 — Package initialization

- [ ] `npm init` with fields from CLAUDE.md (name, version 0.1.0, type module, engines, bin, etc.)
- [ ] Install production dependencies: `better-sqlite3`, `commander`, `zod`, `@modelcontextprotocol/sdk`, `chalk`, `cli-table3`, `date-fns`, `chrono-node`, `uuid`
- [ ] Install dev dependencies: `typescript`, `@types/better-sqlite3`, `@types/uuid`, `tsx`, `vitest`, `eslint`, `typescript-eslint`, `@eslint/js`, `eslint-config-prettier`, `prettier`, `husky`, `lint-staged`, `commitlint`, `@commitlint/config-conventional`
- [ ] Create `tsconfig.json` per CLAUDE.md spec (ES2022, NodeNext, strict, etc.)
- [ ] Create `eslint.config.js` with flat config per CLAUDE.md
- [ ] Create `.prettierrc` with project settings
- [ ] Create `vitest.config.ts` with `pool: 'forks'`, unit/integration projects
- [ ] Create `.gitignore`, `.editorconfig`, `.nvmrc`
- [ ] Set up Husky + lint-staged + commitlint
- [ ] Add all npm scripts from CLAUDE.md to package.json
- [ ] Verify: `npm run build`, `npm run lint`, `npm run test` all pass (empty/trivial)

### 0.2 — CI/CD

- [ ] Create `.github/workflows/ci.yml` — lint, typecheck, test, build on Node 20+22
- [ ] Create `.github/workflows/release-please.yml`
- [ ] Create `.github/workflows/publish.yml`
- [ ] Verify CI passes on push

### 0.3 — Entry point skeleton

- [ ] Create `src/cli/index.ts` — commander program with global flags (`--format`, `--quiet`, `--verbose`, `--db`, `--no-color`), version from package.json, top-level error boundary
- [ ] Create `src/models/errors.ts` — `CliError`, `NotFoundError`, `ValidationError`, `ConflictError`, `DatabaseError`
- [ ] Create `src/formatters/index.ts` — format dispatcher (table/json/csv/tsv based on flag + TTY detection)
- [ ] Verify: `npm run dev -- --help` shows the program help

---

## Step 1: Database Layer

The foundation everything else builds on.

### 1.1 — Connection + migration runner

- [ ] Create `src/db/index.ts`:
  - `getDb(dbPath?)` — singleton that opens `~/.crm/crm.db` (or custom path), sets all pragmas (WAL, busy_timeout, foreign_keys, etc.), runs pending migrations, registers signal handlers for cleanup
  - Auto-create `~/.crm/` directory if it doesn't exist
  - Migration runner: read SQL files from `src/db/migrations/`, compare against `PRAGMA user_version`, execute pending ones in a transaction, update `user_version`
- [ ] Create `src/db/migrations/001_initial.sql` — complete schema for all v1 entities (see 1.2)
- [ ] Test: migration runner against `:memory:` database

### 1.2 — Schema (001_initial.sql)

All tables in one initial migration. This is the full v1 schema.

```sql
-- People
CREATE TABLE people (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  first_name TEXT NOT NULL,
  last_name TEXT,
  email TEXT,
  phone TEXT,
  title TEXT,
  company TEXT,
  location TEXT,
  linkedin TEXT,
  twitter TEXT,
  website TEXT,
  notes TEXT,
  summary TEXT,
  org_id INTEGER REFERENCES organizations(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Organizations
CREATE TABLE organizations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (...),
  name TEXT NOT NULL,
  domain TEXT,
  industry TEXT,
  location TEXT,
  notes TEXT,
  summary TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Interactions
CREATE TABLE interactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (...),
  type TEXT NOT NULL CHECK (type IN ('call', 'email', 'meeting', 'note', 'message')),
  subject TEXT,
  content TEXT,
  direction TEXT CHECK (direction IN ('inbound', 'outbound')),
  occurred_at TEXT NOT NULL DEFAULT (datetime('now')),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Interaction-person junction (many-to-many)
CREATE TABLE interaction_people (
  interaction_id INTEGER NOT NULL REFERENCES interactions(id),
  person_id INTEGER NOT NULL REFERENCES people(id),
  PRIMARY KEY (interaction_id, person_id)
);

-- Deals
CREATE TABLE deals (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (...),
  title TEXT NOT NULL,
  value REAL,
  currency TEXT DEFAULT 'USD',
  stage TEXT NOT NULL DEFAULT 'lead' CHECK (stage IN ('lead', 'prospect', 'proposal', 'negotiation', 'won', 'lost')),
  person_id INTEGER REFERENCES people(id),
  org_id INTEGER REFERENCES organizations(id),
  closed_at TEXT,
  notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Tasks
CREATE TABLE tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE DEFAULT (...),
  title TEXT NOT NULL,
  description TEXT,
  due_at TEXT,
  priority TEXT DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
  completed INTEGER NOT NULL DEFAULT 0,
  completed_at TEXT,
  person_id INTEGER REFERENCES people(id),
  deal_id INTEGER REFERENCES deals(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  archived INTEGER NOT NULL DEFAULT 0
);

-- Tags (polymorphic)
CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE taggings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tag_id INTEGER NOT NULL REFERENCES tags(id),
  entity_type TEXT NOT NULL CHECK (entity_type IN ('person', 'organization', 'deal', 'interaction')),
  entity_id INTEGER NOT NULL,
  UNIQUE(tag_id, entity_type, entity_id)
);

-- Custom fields (polymorphic)
CREATE TABLE custom_fields (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  entity_type TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  field_name TEXT NOT NULL,
  field_value TEXT,
  UNIQUE(entity_type, entity_id, field_name)
);

-- Relationships (person-to-person)
CREATE TABLE relationships (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  person_id INTEGER NOT NULL REFERENCES people(id),
  related_person_id INTEGER NOT NULL REFERENCES people(id),
  type TEXT NOT NULL CHECK (type IN ('colleague', 'friend', 'manager', 'report', 'mentor', 'mentee', 'referred-by', 'referred')),
  notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(person_id, related_person_id, type)
);

-- FTS5 indexes
CREATE VIRTUAL TABLE people_fts USING fts5(first_name, last_name, email, company, notes, summary, content=people, content_rowid=id);
CREATE VIRTUAL TABLE organizations_fts USING fts5(name, domain, industry, notes, summary, content=organizations, content_rowid=id);
CREATE VIRTUAL TABLE interactions_fts USING fts5(subject, content, content=interactions, content_rowid=id);
CREATE VIRTUAL TABLE deals_fts USING fts5(title, notes, content=deals, content_rowid=id);

-- FTS triggers (people — repeat pattern for orgs, interactions, deals)
CREATE TRIGGER people_ai AFTER INSERT ON people BEGIN
  INSERT INTO people_fts(rowid, first_name, last_name, email, company, notes, summary) VALUES (new.id, new.first_name, new.last_name, new.email, new.company, new.notes, new.summary);
END;
CREATE TRIGGER people_ad AFTER DELETE ON people BEGIN
  INSERT INTO people_fts(people_fts, rowid, first_name, last_name, email, company, notes, summary) VALUES ('delete', old.id, old.first_name, old.last_name, old.email, old.company, old.notes, old.summary);
END;
CREATE TRIGGER people_au AFTER UPDATE ON people BEGIN
  INSERT INTO people_fts(people_fts, rowid, first_name, last_name, email, company, notes, summary) VALUES ('delete', old.id, old.first_name, old.last_name, old.email, old.company, old.notes, old.summary);
  INSERT INTO people_fts(rowid, first_name, last_name, email, company, notes, summary) VALUES (new.id, new.first_name, new.last_name, new.email, new.company, new.notes, new.summary);
END;
-- (same pattern for organizations_fts, interactions_fts, deals_fts)

-- Indexes
CREATE INDEX idx_people_org_id ON people(org_id) WHERE archived = 0;
CREATE INDEX idx_people_email ON people(email) WHERE archived = 0;
CREATE INDEX idx_interactions_occurred_at ON interactions(occurred_at);
CREATE INDEX idx_interaction_people_person_id ON interaction_people(person_id);
CREATE INDEX idx_deals_stage ON deals(stage) WHERE archived = 0;
CREATE INDEX idx_deals_person_id ON deals(person_id) WHERE archived = 0;
CREATE INDEX idx_tasks_due_at ON tasks(due_at) WHERE completed = 0 AND archived = 0;
CREATE INDEX idx_tasks_person_id ON tasks(person_id) WHERE archived = 0;
CREATE INDEX idx_taggings_entity ON taggings(entity_type, entity_id);
CREATE INDEX idx_custom_fields_entity ON custom_fields(entity_type, entity_id);
```

### 1.3 — Zod schemas

- [ ] Create `src/models/types.ts` with Row, Insert, and Update schemas for all entities:
  - `PersonRow`, `PersonInsert`, `PersonUpdate`
  - `OrganizationRow`, `OrganizationInsert`, `OrganizationUpdate`
  - `InteractionRow`, `InteractionInsert`
  - `DealRow`, `DealInsert`, `DealUpdate`
  - `TaskRow`, `TaskInsert`, `TaskUpdate`
  - `TagRow`, `TaggingRow`
  - `CustomFieldRow`
  - `RelationshipRow`, `RelationshipInsert`
- [ ] Export inferred TypeScript types for all schemas

---

## Step 2: Formatters

Build the output layer before commands so every command can use it from the start.

### 2.1 — Formatter infrastructure

- [ ] Create `src/formatters/json.ts` — `formatJson(data)`: pretty JSON for TTY, compact for pipe
- [ ] Create `src/formatters/csv.ts` — `formatCsv(data, columns)`: CSV with header row, proper quoting/escaping
- [ ] Create `src/formatters/table.ts` — `formatTable(data, columns)`: cli-table3 with color, truncation to terminal width
- [ ] Create `src/formatters/index.ts` — `formatOutput(data, format, columns)`: dispatcher that picks formatter based on format string. Also handles `--quiet` (output just IDs, one per line) and TSV.
- [ ] Test: formatters as pure functions — input data, assert output strings

---

## Step 3: People (first full vertical slice)

This is the most important step — it establishes the pattern every subsequent entity follows.

### 3.1 — Person repository

- [ ] Create `src/db/repositories/person.repo.ts`:
  - `create(input: PersonInsert): PersonRow`
  - `findById(id: number): PersonRow | null`
  - `findAll(filters?: { tag?, org?, search?, sort?, limit?, offset? }): PersonRow[]`
  - `update(id: number, input: PersonUpdate): PersonRow`
  - `archive(id: number): void`
  - `search(query: string): PersonRow[]` — FTS5 search
- [ ] All queries filter `archived = 0` by default
- [ ] Auto-generate UUID on insert
- [ ] Update `updated_at` on every update
- [ ] Test: full CRUD cycle against in-memory DB

### 3.2 — Person CLI commands

- [ ] Create `src/cli/commands/person.ts`:
  - `crm person add <name>` — parse "First Last", create with optional flags (--email, --phone, --title, --org, --tag, --note, --set key=value)
  - `crm person list` — filterable with --tag, --org, --sort (name, created, last-contacted), --limit, --offset
  - `crm person show <id>` — full detail view. --with interactions/deals/tasks for related data
  - `crm person edit <id>` — update any field via flags, --set key=value for custom fields
  - `crm person delete <id>` — soft delete with confirmation prompt (skip with --force)
- [ ] Register in `src/cli/index.ts`
- [ ] All commands respect --format, --quiet
- [ ] Test: integration tests spawning the CLI process

### 3.3 — Custom fields (for people, reusable for all entities)

- [ ] Create `src/db/repositories/custom-field.repo.ts`:
  - `set(entityType, entityId, fieldName, fieldValue): void`
  - `get(entityType, entityId): CustomFieldRow[]`
  - `delete(entityType, entityId, fieldName): void`
- [ ] Wire into person show (display custom fields) and person edit (--set key=value)
- [ ] Test: set, get, overwrite, delete

---

## Step 4: Organizations

### 4.1 — Organization repository

- [ ] Create `src/db/repositories/org.repo.ts`:
  - `create`, `findById`, `findAll`, `update`, `archive`, `search` — same pattern as person
  - `findByIdWithPeople(id)` — join people where org_id matches
- [ ] Test: CRUD + people association

### 4.2 — Organization CLI commands

- [ ] Create `src/cli/commands/org.ts`:
  - `crm org add <name>` — with --domain, --industry, --location, --note
  - `crm org list` — with --search
  - `crm org show <id>` — with --with people
  - `crm org edit <id>` — update fields
  - `crm org delete <id>` — soft delete
- [ ] Wire person add/edit --org to link person to org (accept org name or ID)
- [ ] Test: CRUD + person-org linkage

---

## Step 5: Interactions

### 5.1 — Interaction repository

- [ ] Create `src/db/repositories/interaction.repo.ts`:
  - `create(input, personIds: number[]): InteractionRow` — insert interaction + junction rows in transaction
  - `findById(id)`, `findByPersonId(personId, filters?)`, `findByOrgId(orgId, filters?)`
  - `findAll(filters?: { type?, since?, until?, personId?, limit? })`
- [ ] Test: create with multiple people, query by person, query by date range

### 5.2 — Interaction CLI commands (crm log)

- [ ] Create `src/cli/commands/interaction.ts`:
  - `crm log <type> <person_ids...>` — type is call/email/meeting/note/message
  - Flags: --subject, --content, --direction, --at (natural language date via chrono-node)
  - `crm log note <id>` with no --content opens `$EDITOR` (or reads stdin if piped)
  - Multiple person IDs for meetings: `crm log meeting 42 43 44`
- [ ] Create `src/utils/editor.ts` — open temp file in $EDITOR, read content back
- [ ] Create `src/utils/dates.ts` — `parseUserDate()` with chrono-node + ISO fallback
- [ ] Test: log various interaction types, verify stdin piping, date parsing

---

## Step 6: Tags

### 6.1 — Tag repository

- [ ] Create `src/db/repositories/tag.repo.ts`:
  - `apply(entityType, entityId, tagName): void` — create tag if not exists, create tagging
  - `remove(entityType, entityId, tagName): void`
  - `findByEntity(entityType, entityId): string[]`
  - `findAll(): TagRow[]` — list all tags with counts
  - `findEntities(tagName, entityType?): number[]` — find entity IDs by tag
- [ ] Test: apply, remove, idempotent apply, list with counts

### 6.2 — Tag CLI commands

- [ ] Create `src/cli/commands/tag.ts`:
  - `crm tag list` — all tags with entity counts
  - `crm tag apply <entity_type> <entity_id> <tag_name>`
  - `crm tag remove <entity_type> <entity_id> <tag_name>`
- [ ] Wire tags into person/org show (display tags) and person/org list (filter by --tag)
- [ ] Wire `crm person add --tag` to auto-apply tag on create
- [ ] Test: full tag lifecycle, filtering by tag

---

## Step 7: Relationships

### 7.1 — Relationship repository

- [ ] Create `src/db/repositories/relationship.repo.ts`:
  - `create(personId, relatedPersonId, type, notes?): void` — create with reciprocal (colleague↔colleague, manager↔report, mentor↔mentee, referred-by↔referred)
  - `findByPersonId(personId): RelationshipRow[]` — with related person name
  - `delete(personId, relatedPersonId, type): void`
- [ ] Test: create, reciprocal creation, query, delete

### 7.2 — Relationship CLI command

- [ ] Add `crm person relate <id> <related_id> --type <type>` subcommand to person.ts
- [ ] Wire into person show (display relationships)
- [ ] Test: relate, show with relationships

---

## Step 8: Deals

### 8.1 — Deal repository

- [ ] Create `src/db/repositories/deal.repo.ts`:
  - `create`, `findById`, `findAll(filters?: { stage?, personId?, orgId? })`, `update`, `archive`
  - `pipeline()` — aggregate by stage (count + total value per stage)
- [ ] Test: CRUD, stage transitions, pipeline aggregation

### 8.2 — Deal CLI commands

- [ ] Create `src/cli/commands/deal.ts`:
  - `crm deal add <title>` — with --value, --person, --org, --stage, --currency
  - `crm deal list` — with --stage (comma-separated filter)
  - `crm deal show <id>`
  - `crm deal edit <id>` — with --stage, --value, --closed-at
  - `crm deal delete <id>`
  - `crm deal pipeline` — table of stages with count and total value
- [ ] Test: CRUD, pipeline output

---

## Step 9: Tasks

### 9.1 — Task repository

- [ ] Create `src/db/repositories/task.repo.ts`:
  - `create`, `findById`, `findAll(filters?: { overdue?, personId?, dealId?, priority? })`
  - `complete(id)` — set completed=1, completed_at=now
  - `archive(id)`
- [ ] `overdue` filter: `due_at < datetime('now') AND completed = 0`
- [ ] Test: CRUD, completion, overdue filtering

### 9.2 — Task CLI commands

- [ ] Create `src/cli/commands/task.ts`:
  - `crm task add <title>` — with --person, --deal, --due (natural language), --priority
  - `crm task list` — with --overdue, --person, --priority
  - `crm task show <id>`
  - `crm task done <id>` — mark complete
  - `crm task delete <id>`
- [ ] Test: CRUD, due date parsing, overdue list

---

## Step 10: Search

### 10.1 — Cross-entity search

- [ ] Create `src/db/repositories/search.repo.ts`:
  - `search(query, filters?: { type?, since? })` — query all FTS5 tables, return unified results with entity type, ID, matched text, rank
  - Weight results: people > orgs > deals > interactions
- [ ] Test: search across entities, type filtering

### 10.2 — Search CLI command

- [ ] Create `src/cli/commands/search.ts`:
  - `crm search <query>` — with --type (filter to entity type), --since, --limit
  - Output: unified table with Type | ID | Name/Title | Match snippet
- [ ] Test: cross-entity search, type filtering

---

## Step 11: Context & Status

### 11.1 — Context command

- [ ] Create `src/cli/commands/context.ts`:
  - `crm context <name_or_id>` — accepts name (fuzzy match via FTS) or integer ID
  - Aggregates: person profile + org + all custom fields + tags + relationships + recent interactions (last 10) + open deals + open tasks + interaction frequency (this week/month/quarter/all-time)
  - JSON output is a single object with all sections
  - Table output is a rich, multi-section formatted display
- [ ] Create `src/formatters/context.ts` — special formatter for the context briefing
- [ ] Test: full context output with all related data

### 11.2 — Status command

- [ ] Create `src/cli/commands/status.ts`:
  - `crm status` — dashboard summary:
    - Total contacts | organizations | open deals (total value)
    - Overdue tasks | interactions this week
    - Recently added contacts (last 7 days)
- [ ] Test: status with various data states

---

## Step 12: MCP Server

### 12.1 — Server setup

- [ ] Create `src/mcp/server.ts`:
  - Initialize `McpServer` with name + version
  - `StdioServerTransport` for Claude Code / other clients
  - Register all tools from tools.ts
  - Graceful shutdown on SIGINT/SIGTERM
- [ ] Create `src/cli/commands/mcp.ts`:
  - `crm mcp serve` — start the MCP server (stdio)

### 12.2 — MCP tools

- [ ] Create `src/mcp/tools.ts` — register all tools:
  - `crm_person_search` — search people by name, email, tag, org
  - `crm_person_get` — full person detail (same as `crm person show`)
  - `crm_person_create` — create a person
  - `crm_person_update` — update person fields
  - `crm_person_update_summary` — update AI-maintained summary
  - `crm_org_search` — search organizations
  - `crm_org_get` — org detail with people
  - `crm_interaction_log` — log an interaction
  - `crm_interaction_list` — list interactions for a person/org
  - `crm_search` — cross-entity FTS search
  - `crm_context` — full person context briefing
  - `crm_task_create` — create a task
  - `crm_task_list` — list open/overdue tasks
  - `crm_deal_create` — create a deal
  - `crm_deal_update` — update deal stage/value
  - `crm_tag_apply` — apply a tag
  - `crm_stats` — CRM summary stats (same as `crm status`)
- [ ] Each tool: Zod input schema, calls repository, returns JSON content
- [ ] Errors returned as `{ content: [...], isError: true }`, never thrown
- [ ] Test: MCP tools via inspector or programmatic client

---

## Step 13: Polish & Ship

### 13.1 — Final integration

- [ ] End-to-end smoke test: full workflow (add person → add org → link → log interaction → add deal → add task → search → context)
- [ ] Verify all output formats (table, json, csv, tsv, quiet) work correctly for every command
- [ ] Verify piping works: `crm person list -q | xargs -I{} crm person show {}`
- [ ] Verify exit codes are correct for all error conditions
- [ ] Verify `--db` flag works for alternate database paths
- [ ] Verify `--no-color` disables chalk output

### 13.2 — Documentation

- [ ] Update README.md if any commands changed during implementation
- [ ] Update CLAUDE.md if any patterns changed
- [ ] Add LICENSE (MIT)

### 13.3 — Release

- [ ] Verify `npm run check` passes (lint + typecheck + test)
- [ ] Verify `npm pack --dry-run` contains only intended files
- [ ] Tag as v0.1.0 via release-please
- [ ] Publish to npm

---

## Execution Order & Dependencies

```
Step 0  (scaffolding)
  │
Step 1  (database + schema + types)
  │
Step 2  (formatters)
  │
Step 3  (people) ─── establishes the pattern
  │
  ├── Step 4  (organizations) ── depends on people for org linkage
  │
  ├── Step 5  (interactions) ── depends on people
  │
  ├── Step 6  (tags) ── depends on people/orgs existing
  │
  └── Step 7  (relationships) ── depends on people
  │
  ├── Step 8  (deals) ── depends on people/orgs
  │
  └── Step 9  (tasks) ── depends on people/deals
  │
Step 10 (search) ── depends on all entities existing
  │
Step 11 (context + status) ── depends on all entities + search
  │
Step 12 (MCP server) ── depends on all repositories
  │
Step 13 (polish + ship)
```

Steps 4-9 can be parallelized after Step 3 establishes the pattern. Steps 4-7 and 8-9 are two natural groups.

---

## Estimated Scope

- ~15 source files (repos, commands, formatters, utils, MCP)
- ~1 migration file (initial schema)
- ~10 test files (unit + integration)
- Total: roughly 3,000-5,000 lines of TypeScript
