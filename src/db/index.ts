import { existsSync, mkdirSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import Database from 'better-sqlite3'

const __dirname = dirname(fileURLToPath(import.meta.url))
const MIGRATIONS_DIR = join(__dirname, 'migrations')
const DEFAULT_DB_DIR = join(process.env['HOME'] ?? '~', '.crm')
const DEFAULT_DB_PATH = join(DEFAULT_DB_DIR, 'crm.db')

let db: Database.Database | null = null

function applyPragmas(database: Database.Database): void {
  database.pragma('journal_mode = WAL')
  database.pragma('busy_timeout = 5000')
  database.pragma('synchronous = NORMAL')
  database.pragma('foreign_keys = ON')
  database.pragma('cache_size = -64000')
  database.pragma('temp_store = MEMORY')
}

function parseMigration(sql: string): { up: string; down: string } {
  const upMatch = /--\s*up\s*\n([\s\S]*?)(?=--\s*down|$)/i.exec(sql)
  const downMatch = /--\s*down\s*\n([\s\S]*?)$/i.exec(sql)
  return {
    up: upMatch?.[1]?.trim() ?? '',
    down: downMatch?.[1]?.trim() ?? '',
  }
}

function getMigrationFiles(): { version: number; filename: string }[] {
  if (!existsSync(MIGRATIONS_DIR)) return []

  return readdirSync(MIGRATIONS_DIR)
    .filter((f) => f.endsWith('.sql'))
    .sort()
    .map((filename) => {
      const match = /^(\d+)_/.exec(filename)
      const version = match?.[1] ? parseInt(match[1], 10) : 0
      return { version, filename }
    })
}

export function runMigrations(database: Database.Database): void {
  const migrations = getMigrationFiles()
  const currentVersion = database.pragma('user_version', { simple: true }) as number

  const pending = migrations.filter((m) => m.version > currentVersion)
  if (pending.length === 0) return

  const migrate = database.transaction(() => {
    for (const migration of pending) {
      const sql = readFileSync(join(MIGRATIONS_DIR, migration.filename), 'utf-8')
      const { up } = parseMigration(sql)
      if (up) {
        database.exec(up)
      }
    }
    const lastVersion = pending[pending.length - 1]?.version ?? currentVersion
    database.pragma(`user_version = ${String(lastVersion)}`)
  })

  migrate()
}

export function getDb(dbPath?: string): Database.Database {
  if (db) return db

  const resolvedPath = dbPath ?? DEFAULT_DB_PATH

  if (resolvedPath !== ':memory:') {
    const dir = dirname(resolvedPath)
    if (!existsSync(dir)) {
      mkdirSync(dir, { recursive: true })
    }
  }

  db = new Database(resolvedPath)
  applyPragmas(db)
  runMigrations(db)

  return db
}

export function createMemoryDb(): Database.Database {
  const memDb = new Database(':memory:')
  applyPragmas(memDb)
  runMigrations(memDb)
  return memDb
}

export function closeDb(): void {
  if (db) {
    db.close()
    db = null
  }
}

function setupSignalHandlers(): void {
  const cleanup = () => {
    closeDb()
  }

  process.on('exit', cleanup)
  process.on('SIGINT', () => {
    cleanup()
    process.exit(130)
  })
  process.on('SIGTERM', () => {
    cleanup()
    process.exit(143)
  })
}

setupSignalHandlers()
