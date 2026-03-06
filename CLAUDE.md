# CLAUDE.md — crm-cli

## Project Overview

`crm` is a local-first personal CRM for the terminal. TypeScript, SQLite (`better-sqlite3`), commander.js CLI, built-in MCP server. See `README.md` for full product spec.

## Quick Reference

- **Language:** TypeScript (strict mode, ESM-only)
- **Runtime:** Node.js >= 20
- **Package manager:** npm
- **Database:** SQLite via `better-sqlite3` (WAL mode)
- **CLI framework:** commander
- **Validation:** zod
- **Testing:** vitest
- **Build:** tsc (production), tsx (development)
- **Linting:** ESLint v9+ (flat config) + Prettier
- **Releases:** release-please + conventional commits

## Commands

```bash
npm run dev             # Run CLI in development (tsx)
npm run build           # Compile with tsc
npm run test            # Run all tests (vitest)
npm run test:unit       # Unit tests only
npm run test:int        # Integration tests only
npm run test:watch      # Watch mode
npm run lint            # ESLint + Prettier check
npm run lint:fix        # Auto-fix lint issues
npm run typecheck       # tsc --noEmit
npm run check           # lint + typecheck + test (CI pipeline locally)
```

## Project Structure

```
src/
  cli/
    index.ts              # Entry point, commander setup, global flags
    commands/             # One file per entity (person.ts, org.ts, etc.)
  db/
    index.ts              # Connection singleton, migration runner, pragmas
    migrations/           # Sequential SQL files (001_initial.sql, ...)
    repositories/         # Data access layer (one repo per entity)
  mcp/
    server.ts             # MCP server setup + stdio transport
    tools.ts              # MCP tool definitions
  formatters/             # Output formatting (table, json, csv)
  models/
    types.ts              # Zod schemas + inferred TS types
  utils/                  # Dates, config, editor helpers
tests/
  unit/                   # Fast, no I/O, in-memory DB
  integration/            # Spawn CLI process, assert stdout/stderr/exit code
  fixtures/               # Test data
```

## Architecture Rules

### Layers

1. **CLI commands** (`src/cli/commands/`) — parse args, call repositories, format output. No direct SQL.
2. **Repositories** (`src/db/repositories/`) — all database queries. Accept and return typed objects. Use prepared statements.
3. **Formatters** (`src/formatters/`) — transform data to table/json/csv. No business logic.
4. **Models** (`src/models/types.ts`) — Zod schemas are the single source of truth for types. Infer TS types with `z.infer<>`.
5. **MCP tools** (`src/mcp/`) — thin wrappers around the same repositories the CLI uses. No duplicate logic.

### Key Patterns

- **Single DB connection** — `src/db/index.ts` exports a singleton. WAL mode + pragmas on open. Close on process exit via signal handlers.
- **Prepared statements** — always use parameterized queries (`db.prepare(...).run(params)`). Never interpolate user input into SQL.
- **Transactions** — wrap multi-step mutations in `db.transaction(...)`.
- **Soft deletes** — set `archived = 1`, never `DELETE FROM`. Filter `archived = 0` by default in all queries.
- **Integer IDs + UUIDs** — integer `id` for CLI/human use, `uuid` column for external/API references.
- **Exit codes** — 0=success, 1=error, 2=usage error, 3=not found, 4=conflict, 10=db error. Use `process.exitCode`, not `process.exit()` in command handlers (allows cleanup).
- **stdout/stderr discipline** — data output (tables, JSON, CSV) goes to `stdout` only. All human messages (success confirmations, progress, warnings, errors) go to `stderr`. This is critical for piping.
- **Output format** — respect `--format` flag everywhere. Default to table for TTY, JSON for pipes. Check `process.stdout.isTTY`.
- **Error handling** — use a `CliError` class hierarchy with exit codes. Catch at top level, print human-readable message to stderr (prefixed `crm: error:`), set exit code. Show stack traces only with `--verbose`.
- **Signal handling** — handle `SIGINT`/`SIGTERM` to close DB connection before exit. Exit with `128 + signal_number` (SIGINT=130, SIGTERM=143).

### Error Hierarchy

```
CliError (exitCode: 1)
├── ValidationError (exitCode: 2)
├── NotFoundError (exitCode: 3)
├── ConflictError (exitCode: 4)
└── DatabaseError (exitCode: 10)
```

Top-level error boundary wraps `program.parseAsync()` — translates errors to exit codes and user-facing messages. Catch `ZodError` separately and format as user-friendly field-level messages.

## TypeScript Configuration

- **Target:** ES2022 (Node 20+)
- **Module:** `NodeNext` (ESM with `.js` extensions in imports)
- **Module resolution:** `NodeNext`
- **Strict mode:** `strict: true` + `noUncheckedIndexedAccess: true` + `exactOptionalProperties: true`
- **Other flags:** `skipLibCheck: true`, `declaration: true`, `sourceMap: true`, `outDir: "./dist"`, `rootDir: "./src"`
- **File extensions:** always use `.js` in import paths (e.g., `import { foo } from './bar.js'`) — required for Node ESM
- **No enums** — use `as const` objects or union types instead
- **No `any`** — use `unknown` and narrow with type guards or Zod

## Database Conventions

### Connection Setup

Enable these pragmas immediately after opening the database:

```sql
PRAGMA journal_mode = WAL;          -- concurrent reads during writes
PRAGMA busy_timeout = 5000;         -- wait up to 5s instead of SQLITE_BUSY
PRAGMA synchronous = NORMAL;        -- safe with WAL, faster than FULL
PRAGMA foreign_keys = ON;           -- enforce FK constraints (off by default!)
PRAGMA cache_size = -64000;         -- 64MB cache
PRAGMA temp_store = MEMORY;         -- temp tables in memory
```

### Migrations

- Sequential SQL files in `src/db/migrations/`. Name format: `NNN_description.sql`
- Each migration file has `-- up` and `-- down` sections
- Track current version via `PRAGMA user_version` — compare against available migration files on startup
- Run migrations inside `db.transaction(...)` for atomicity
- Auto-backup `crm.db` before running migrations
- **Never modify existing migrations** — always create a new one

### Schema

- **FTS5** indexes on people, organizations, interactions, deals — keep FTS tables in sync via triggers
- **Timestamps** — store as ISO 8601 strings (`TEXT`). Use `datetime('now')` for defaults.
- **Booleans** — `INTEGER` (0/1) since SQLite has no native boolean
- Use `.pluck(true)` for single-column queries, `.expand(true)` for JOINs with nested objects

## Zod Schema Patterns

Define a base row schema, then derive insert/update variants:

```typescript
const PersonRow = z.object({
  id: z.number(),
  uuid: z.string().uuid(),
  name: z.string(),
  email: z.string().email().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  archived: z.number().transform(v => v === 1),
})

type PersonRow = z.infer<typeof PersonRow>

const PersonInsert = PersonRow.omit({ id: true, uuid: true, created_at: true, updated_at: true, archived: true })
const PersonUpdate = PersonInsert.partial()  // all fields optional for PATCH semantics
```

For CLI input: use `z.coerce.number()` for string-to-number args, `z.enum()` for fixed choices, `.default()` for optional args.

## Testing

- **Framework:** vitest with `pool: 'forks'` (required — `better-sqlite3` native addon doesn't work with worker threads)
- **Unit tests** (`tests/unit/`) — test repositories, formatters, utils in isolation. Use `:memory:` SQLite databases.
- **Integration tests** (`tests/integration/`) — spawn the CLI process, assert on stdout, stderr, and exit code. Use `--db /tmp/test-xxx/crm.db` for isolation.
- **Fixtures** — shared test data in `tests/fixtures/`
- **Test naming:** `describe('PersonRepository')` / `it('should find person by email')`
- **Each test gets a fresh DB** — create in-memory database, run migrations, seed, test, discard. Setup is instant because `better-sqlite3` is synchronous.
- **No mocking the database** in integration tests — use real SQLite in-memory
- **Test output formats** — verify JSON output parses correctly, table output contains expected strings
- **Commander testability** — use `program.exitOverride()` to throw instead of `process.exit()` during unit tests. Capture output via `vi.spyOn(process.stdout, 'write')`.
- **Date testing** — always pass a reference date to `chrono-node` for deterministic results
- **Snapshot testing** — use `toMatchInlineSnapshot()` for table formatting output

## Code Style

### ESLint (v9+ flat config)

Use `eslint.config.js` with `typescript-eslint` v8+:

```javascript
import eslint from '@eslint/js'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  { ignores: ['dist/', 'coverage/'] },
)
```

Use `eslint-config-prettier` (last in config) to disable rules that conflict with Prettier. Do NOT use `eslint-plugin-prettier`.

### Prettier

```json
{ "singleQuote": true, "trailingComma": "all", "semi": false, "printWidth": 100 }
```

### Naming

- Files: `kebab-case.ts` (e.g., `person.repo.ts`, `cli-table.ts`)
- Types/interfaces: `PascalCase` (e.g., `Person`, `CreatePersonInput`)
- Functions/variables: `camelCase`
- Constants: `UPPER_SNAKE_CASE` for true constants, `camelCase` for derived values
- Database columns: `snake_case`
- CLI flags: `--kebab-case`

### Other

- **Imports** — group: node builtins → external packages → internal modules → types. Sorted alphabetically within groups.
- **No default exports** — use named exports everywhere
- **Prefer `const` arrow functions** for module-level functions, regular functions for exported/hoisted functions

## Git Conventions

### Commits

All commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(person): add merge command
fix(db): handle WAL lock contention
docs: update README with new examples
chore(deps): update better-sqlite3 to v11
refactor(repo): extract base repository class
test(deal): add pipeline stage transition tests
feat(cli)!: rename log command to interact    # breaking change
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`

Enforced via `commitlint` + `@commitlint/config-conventional` in a git hook.

### Branches

```
feat/add-person-merge
fix/wal-lock-contention
docs/contributing-guide
refactor/repository-pattern
```

### PR Workflow

- All work on feature branches off `main`
- `main` is always releasable
- **Squash merge** into `main` (required — release-please parses commits on main)
- Delete branches after merge
- CI must pass before merge

## Versioning & Releases

- **Semantic Versioning** — `MAJOR.MINOR.PATCH`. Start at `0.1.0`, ship `1.0.0` when the public API (command names, flags, exit codes, JSON output shape, MCP tool signatures) is stable.
- **release-please** — GitHub Action auto-creates release PRs from conventional commits, bumps version in `package.json`, generates `CHANGELOG.md`, creates GitHub Release.
- **npm publishing** — automated via GitHub Actions when release-please PR is merged. Use `npm publish --provenance --access public` for supply chain security.

## CI/CD (GitHub Actions)

Three workflow files:

### `ci.yml` — every push and PR

1. **Lint** — `npm run lint` + `npm run typecheck`
2. **Test** — `npm run test` on Node 20 and 22 (matrix)
3. **Build** — `npm run build`

Use `npm ci` (not `npm install`) for reproducible builds. Cache via `setup-node`'s `cache: 'npm'`.

### `release-please.yml` — push to main

Runs `googleapis/release-please-action@v4` with `release-type: node`. Creates/updates a release PR automatically.

### `publish.yml` — on GitHub Release published

Runs `npm publish --provenance --access public`. Requires `permissions: id-token: write` for provenance and `NPM_TOKEN` secret.

## Package Configuration

```jsonc
{
  "name": "crm-cli",
  "version": "0.1.0",
  "type": "module",
  "bin": { "crm": "./dist/cli/index.js" },
  "exports": { ".": { "import": "./dist/index.js", "types": "./dist/index.d.ts" } },
  "engines": { "node": ">=20" },
  "files": ["dist", "LICENSE", "README.md"],
  "publishConfig": { "access": "public" },
  "scripts": {
    "dev": "tsx src/cli/index.ts",
    "build": "tsc",
    "test": "vitest run",
    "lint": "eslint . && prettier --check .",
    "typecheck": "tsc --noEmit",
    "prepublishOnly": "npm run build",
    "prepare": "husky"
  }
}
```

- The `bin` entry point needs `#!/usr/bin/env node` shebang
- `files` limits what's published — never ship `src/`, `tests/`, config files
- `exports` field exposes the MCP server for programmatic use
- Verify with `npm pack --dry-run` before publishing

## Build

Use **`tsc`** for production builds (not tsup/esbuild) because `better-sqlite3` is a native addon that cannot be bundled. The compiled output mirrors the `src/` structure in `dist/`. This is simpler and more debuggable.

Use **`tsx`** for development — runs TypeScript directly via esbuild transpilation, supports ESM, very fast.

## MCP Server

- Uses `@modelcontextprotocol/sdk` with `StdioServerTransport`
- **Never write to stdout** in MCP mode — stdout is the transport channel. All logging goes to `stderr`.
- Tool definitions use Zod schemas for input validation (reuse schemas from `types.ts`)
- Each MCP tool is a thin wrapper calling the same repository methods as CLI commands
- Return errors as MCP content with `isError: true` — do not throw exceptions from tool handlers
- Test MCP tools with `@modelcontextprotocol/inspector`
- Handle graceful shutdown: listen for `SIGINT`/`SIGTERM`, call `server.close()`

## Date Handling

Use `chrono-node` for natural language date parsing with a fallback chain:

1. Try chrono-node (natural language: "next friday", "2 weeks ago")
2. Fall back to ISO 8601 parsing (`new Date(input)`)
3. Error with helpful message if both fail

Always pass a reference date to chrono for deterministic behavior (critical for tests). Store all dates as ISO 8601 UTC strings in SQLite. Convert to local time at display.

Integrate into Zod schemas via `.transform()` for CLI input validation.

## Dependencies Policy

- Keep dependencies minimal — startup time matters for a CLI
- `better-sqlite3` is a native module — test against Node 20 and 22 in CI
- Security: run `npm audit` in CI, fail on high/critical vulnerabilities

## Git Hooks (Husky + lint-staged)

Pre-commit: `lint-staged` runs ESLint + Prettier on staged files only.

Commit-msg: `commitlint` enforces conventional commit format.

```json
{
  "lint-staged": {
    "*.{ts,js}": ["eslint --fix", "prettier --write"],
    "*.{json,md,yml,yaml}": ["prettier --write"]
  }
}
```

## Repository Files

- `.gitignore` — `node_modules/`, `dist/`, `*.tsbuildinfo`, `coverage/`, `.env`, `.DS_Store`
- `.editorconfig` — LF line endings, UTF-8, 2-space indent, trim trailing whitespace
- `.nvmrc` — `22` (pin Node.js version for contributors)
- `LICENSE` — MIT
- `CONTRIBUTING.md` — dev setup, commit format, PR process, testing expectations

## Common Tasks

### Adding a new entity

1. Add Zod schemas to `src/models/types.ts` (Row, Insert, Update variants)
2. Create migration in `src/db/migrations/`
3. Create repository in `src/db/repositories/`
4. Create CLI command in `src/cli/commands/`
5. Register command in `src/cli/index.ts`
6. Add MCP tools in `src/mcp/tools.ts`
7. Add tests for repository (unit) + command (integration)

### Adding a new CLI command to an existing entity

1. Add the subcommand in the entity's command file
2. Call the appropriate repository method
3. Format output through the formatter
4. Add unit test for the repository method
5. Add integration test for the CLI command

### Adding a new MCP tool

1. Define Zod input schema (reuse from `types.ts` if possible)
2. Add tool definition in `src/mcp/tools.ts`
3. Wire to existing repository method
4. Test with MCP inspector
