import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createInteractionRepo } from '../../src/db/repositories/interaction.repo.js'

let db: Database.Database

function setup() {
  db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

function seedPeople() {
  db.prepare("INSERT INTO people (first_name) VALUES ('Jane')").run()
  db.prepare("INSERT INTO people (first_name) VALUES ('Bob')").run()
}

describe('InteractionRepository', () => {
  beforeEach(() => {
    db = setup()
    seedPeople()
  })

  it('should create an interaction with person links', () => {
    const repo = createInteractionRepo(db)
    const interaction = repo.create({ type: 'call', subject: 'Catch-up call' }, [1])

    expect(interaction.id).toBe(1)
    expect(interaction.type).toBe('call')
    expect(interaction.subject).toBe('Catch-up call')
    expect(interaction.person_ids).toEqual([1])
    expect(interaction.uuid).toBeTruthy()
    expect(interaction.archived).toBe(0)
  })

  it('should create an interaction with multiple people', () => {
    const repo = createInteractionRepo(db)
    const interaction = repo.create({ type: 'meeting', subject: 'Group meeting' }, [1, 2])

    expect(interaction.person_ids).toEqual([1, 2])
  })

  it('should create an interaction with no people', () => {
    const repo = createInteractionRepo(db)
    const interaction = repo.create({ type: 'note', subject: 'General note' }, [])

    expect(interaction.person_ids).toEqual([])
  })

  it('should find an interaction by ID', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'Test' }, [1])

    const found = repo.findById(1)
    expect(found).toBeTruthy()
    expect(found?.type).toBe('call')
    expect(found?.person_ids).toEqual([1])
  })

  it('should return undefined for missing interaction', () => {
    const repo = createInteractionRepo(db)
    expect(repo.findById(999)).toBeUndefined()
  })

  it('should list all interactions', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'Call 1' }, [1])
    repo.create({ type: 'email', subject: 'Email 1' }, [2])

    const interactions = repo.findAll()
    expect(interactions).toHaveLength(2)
  })

  it('should filter by person ID', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'Call with Jane' }, [1])
    repo.create({ type: 'email', subject: 'Email to Bob' }, [2])

    const results = repo.findAll({ personId: 1 })
    expect(results).toHaveLength(1)
    expect(results[0]?.subject).toBe('Call with Jane')
  })

  it('should filter by type', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'Call 1' }, [1])
    repo.create({ type: 'email', subject: 'Email 1' }, [1])

    const results = repo.findAll({ type: 'email' })
    expect(results).toHaveLength(1)
    expect(results[0]?.type).toBe('email')
  })

  it('should filter by org ID', () => {
    const repo = createInteractionRepo(db)
    db.prepare('INSERT INTO organizations (name) VALUES (?)').run('Acme')
    db.prepare('UPDATE people SET org_id = 1 WHERE id = 1').run()

    repo.create({ type: 'call', subject: 'Acme call' }, [1])
    repo.create({ type: 'call', subject: 'Bob call' }, [2])

    const results = repo.findAll({ orgId: 1 })
    expect(results).toHaveLength(1)
    expect(results[0]?.subject).toBe('Acme call')
  })

  it('should limit results', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'A' }, [1])
    repo.create({ type: 'call', subject: 'B' }, [1])
    repo.create({ type: 'call', subject: 'C' }, [1])

    const results = repo.findAll({ limit: 2 })
    expect(results).toHaveLength(2)
  })

  it('should archive (soft delete) an interaction', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'Test' }, [1])

    const archived = repo.archive(1)
    expect(archived).toBe(true)

    expect(repo.findAll()).toHaveLength(0)
    expect(repo.findById(1)).toBeUndefined()
  })

  it('should return false when archiving nonexistent interaction', () => {
    const repo = createInteractionRepo(db)
    expect(repo.archive(999)).toBe(false)
  })

  it('should search via FTS', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'call', subject: 'Discussed roadmap' }, [1])
    repo.create({ type: 'email', subject: 'Invoice sent' }, [2])

    const results = repo.search('roadmap')
    expect(results).toHaveLength(1)
    expect(results[0]?.subject).toBe('Discussed roadmap')
  })

  it('should not find archived interactions in search', () => {
    const repo = createInteractionRepo(db)
    repo.create({ type: 'note', content: 'Secret note' }, [1])
    repo.archive(1)

    const results = repo.search('Secret')
    expect(results).toHaveLength(0)
  })

  it('should set direction', () => {
    const repo = createInteractionRepo(db)
    const interaction = repo.create(
      { type: 'email', subject: 'Outbound email', direction: 'outbound' },
      [1],
    )

    expect(interaction.direction).toBe('outbound')
  })

  it('should set custom occurred_at', () => {
    const repo = createInteractionRepo(db)
    const interaction = repo.create(
      { type: 'call', subject: 'Past call', occurred_at: '2026-01-15 10:00:00' },
      [1],
    )

    expect(interaction.occurred_at).toBe('2026-01-15 10:00:00')
  })
})
