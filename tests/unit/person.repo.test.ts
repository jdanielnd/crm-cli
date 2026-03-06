import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createPersonRepo } from '../../src/db/repositories/person.repo.js'
import { createCustomFieldRepo } from '../../src/db/repositories/custom-field.repo.js'

let db: Database.Database

function setup() {
  db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

describe('PersonRepository', () => {
  beforeEach(() => {
    db = setup()
  })

  it('should create a person', () => {
    const repo = createPersonRepo(db)
    const person = repo.create({
      first_name: 'Jane',
      last_name: 'Smith',
      email: 'jane@example.com',
    })

    expect(person.id).toBe(1)
    expect(person.first_name).toBe('Jane')
    expect(person.last_name).toBe('Smith')
    expect(person.email).toBe('jane@example.com')
    expect(person.uuid).toBeTruthy()
    expect(person.archived).toBe(0)
  })

  it('should create a person with minimal fields', () => {
    const repo = createPersonRepo(db)
    const person = repo.create({ first_name: 'Jane' })

    expect(person.id).toBe(1)
    expect(person.first_name).toBe('Jane')
    expect(person.last_name).toBeNull()
    expect(person.email).toBeNull()
  })

  it('should find a person by ID', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane' })

    const person = repo.findById(1)
    expect(person).toBeTruthy()
    expect(person?.first_name).toBe('Jane')
  })

  it('should return undefined for missing person', () => {
    const repo = createPersonRepo(db)
    expect(repo.findById(999)).toBeUndefined()
  })

  it('should list all people', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane' })
    repo.create({ first_name: 'Bob' })

    const people = repo.findAll()
    expect(people).toHaveLength(2)
  })

  it('should filter by search', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane', email: 'jane@example.com' })
    repo.create({ first_name: 'Bob', email: 'bob@example.com' })

    const results = repo.findAll({ search: 'jane' })
    expect(results).toHaveLength(1)
    expect(results[0]?.first_name).toBe('Jane')
  })

  it('should sort by name', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Zara' })
    repo.create({ first_name: 'Alice' })

    const people = repo.findAll({ sort: 'name' })
    expect(people[0]?.first_name).toBe('Alice')
    expect(people[1]?.first_name).toBe('Zara')
  })

  it('should limit results', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'A' })
    repo.create({ first_name: 'B' })
    repo.create({ first_name: 'C' })

    const people = repo.findAll({ limit: 2 })
    expect(people).toHaveLength(2)
  })

  it('should update a person', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane', email: 'old@example.com' })

    const updated = repo.update(1, { email: 'new@example.com' })
    expect(updated?.email).toBe('new@example.com')
    expect(updated?.first_name).toBe('Jane')
  })

  it('should return undefined when updating nonexistent person', () => {
    const repo = createPersonRepo(db)
    expect(repo.update(999, { email: 'test@example.com' })).toBeUndefined()
  })

  it('should archive (soft delete) a person', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane' })

    const archived = repo.archive(1)
    expect(archived).toBe(true)

    // Should not appear in findAll
    expect(repo.findAll()).toHaveLength(0)

    // Should not appear in findById
    expect(repo.findById(1)).toBeUndefined()
  })

  it('should return false when archiving nonexistent person', () => {
    const repo = createPersonRepo(db)
    expect(repo.archive(999)).toBe(false)
  })

  it('should search via FTS', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane', last_name: 'Smith', notes: 'Met at conference' })
    repo.create({ first_name: 'Bob', last_name: 'Jones' })

    const results = repo.search('conference')
    expect(results).toHaveLength(1)
    expect(results[0]?.first_name).toBe('Jane')
  })

  it('should not find archived people in search', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane', notes: 'archived person' })
    repo.archive(1)

    const results = repo.search('archived')
    expect(results).toHaveLength(0)
  })

  it('should filter by org ID', () => {
    const repo = createPersonRepo(db)
    db.prepare('INSERT INTO organizations (name) VALUES (?)').run('Acme')

    repo.create({ first_name: 'Jane', org_id: 1 })
    repo.create({ first_name: 'Bob' })

    const results = repo.findAll({ orgId: 1 })
    expect(results).toHaveLength(1)
    expect(results[0]?.first_name).toBe('Jane')
  })

  it('should filter by tag', () => {
    const repo = createPersonRepo(db)
    repo.create({ first_name: 'Jane' })
    repo.create({ first_name: 'Bob' })

    db.prepare("INSERT INTO tags (name) VALUES ('vip')").run()
    db.prepare(
      "INSERT INTO taggings (tag_id, entity_type, entity_id) VALUES (1, 'person', 1)",
    ).run()

    const results = repo.findAll({ tag: 'vip' })
    expect(results).toHaveLength(1)
    expect(results[0]?.first_name).toBe('Jane')
  })
})

describe('CustomFieldRepository', () => {
  beforeEach(() => {
    db = setup()
  })

  it('should set and get custom fields', () => {
    const cfRepo = createCustomFieldRepo(db)
    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')

    cfRepo.set('person', 1, 'birthday', '1990-03-15')
    cfRepo.set('person', 1, 'github', 'janesmith')

    const fields = cfRepo.get('person', 1)
    expect(fields).toHaveLength(2)
    expect(fields.find((f) => f.field_name === 'birthday')?.field_value).toBe('1990-03-15')
    expect(fields.find((f) => f.field_name === 'github')?.field_value).toBe('janesmith')
  })

  it('should overwrite existing custom field', () => {
    const cfRepo = createCustomFieldRepo(db)
    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')

    cfRepo.set('person', 1, 'birthday', '1990-03-15')
    cfRepo.set('person', 1, 'birthday', '1991-04-20')

    const fields = cfRepo.get('person', 1)
    expect(fields).toHaveLength(1)
    expect(fields[0]?.field_value).toBe('1991-04-20')
  })

  it('should delete a custom field', () => {
    const cfRepo = createCustomFieldRepo(db)
    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')

    cfRepo.set('person', 1, 'birthday', '1990-03-15')
    const deleted = cfRepo.delete('person', 1, 'birthday')

    expect(deleted).toBe(true)
    expect(cfRepo.get('person', 1)).toHaveLength(0)
  })

  it('should return false when deleting nonexistent field', () => {
    const cfRepo = createCustomFieldRepo(db)
    expect(cfRepo.delete('person', 1, 'nonexistent')).toBe(false)
  })
})
