import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createRelationshipRepo } from '../../src/db/repositories/relationship.repo.js'

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
  db.prepare("INSERT INTO people (first_name) VALUES ('Alice')").run()
}

describe('RelationshipRepository', () => {
  beforeEach(() => {
    db = setup()
    seedPeople()
  })

  it('should create a relationship', () => {
    const repo = createRelationshipRepo(db)
    const rel = repo.create({
      person_id: 1,
      related_person_id: 2,
      type: 'colleague',
    })

    expect(rel.id).toBe(1)
    expect(rel.person_id).toBe(1)
    expect(rel.related_person_id).toBe(2)
    expect(rel.type).toBe('colleague')
    expect(rel.notes).toBeNull()
  })

  it('should create a relationship with notes', () => {
    const repo = createRelationshipRepo(db)
    const rel = repo.create({
      person_id: 1,
      related_person_id: 2,
      type: 'friend',
      notes: 'Met at conference',
    })

    expect(rel.notes).toBe('Met at conference')
  })

  it('should find relationships for a person', () => {
    const repo = createRelationshipRepo(db)
    repo.create({ person_id: 1, related_person_id: 2, type: 'colleague' })
    repo.create({ person_id: 3, related_person_id: 1, type: 'friend' })

    const rels = repo.findForPerson(1)
    expect(rels).toHaveLength(2)
  })

  it('should find relationships by type', () => {
    const repo = createRelationshipRepo(db)
    repo.create({ person_id: 1, related_person_id: 2, type: 'colleague' })
    repo.create({ person_id: 1, related_person_id: 3, type: 'friend' })

    const colleagues = repo.findByType(1, 'colleague')
    expect(colleagues).toHaveLength(1)
    expect(colleagues[0]?.related_person_id).toBe(2)
  })

  it('should return empty array for no relationships', () => {
    const repo = createRelationshipRepo(db)
    expect(repo.findForPerson(1)).toEqual([])
  })

  it('should delete a relationship', () => {
    const repo = createRelationshipRepo(db)
    repo.create({ person_id: 1, related_person_id: 2, type: 'colleague' })

    const deleted = repo.delete(1)
    expect(deleted).toBe(true)
    expect(repo.findForPerson(1)).toHaveLength(0)
  })

  it('should return false when deleting nonexistent relationship', () => {
    const repo = createRelationshipRepo(db)
    expect(repo.delete(999)).toBe(false)
  })

  it('should enforce unique constraint on person_id, related_person_id, type', () => {
    const repo = createRelationshipRepo(db)
    repo.create({ person_id: 1, related_person_id: 2, type: 'colleague' })

    expect(() => {
      repo.create({ person_id: 1, related_person_id: 2, type: 'colleague' })
    }).toThrow()
  })

  it('should allow same pair with different types', () => {
    const repo = createRelationshipRepo(db)
    repo.create({ person_id: 1, related_person_id: 2, type: 'colleague' })
    repo.create({ person_id: 1, related_person_id: 2, type: 'friend' })

    const rels = repo.findForPerson(1)
    expect(rels).toHaveLength(2)
  })
})
