import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createTagRepo } from '../../src/db/repositories/tag.repo.js'

let db: Database.Database

function setup() {
  db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

function seedEntities() {
  db.prepare("INSERT INTO people (first_name) VALUES ('Jane')").run()
  db.prepare("INSERT INTO people (first_name) VALUES ('Bob')").run()
  db.prepare("INSERT INTO organizations (name) VALUES ('Acme')").run()
}

describe('TagRepository', () => {
  beforeEach(() => {
    db = setup()
    seedEntities()
  })

  it('should list all tags', () => {
    const repo = createTagRepo(db)
    db.prepare("INSERT INTO tags (name) VALUES ('vip')").run()
    db.prepare("INSERT INTO tags (name) VALUES ('client')").run()

    const tags = repo.findAll()
    expect(tags).toHaveLength(2)
    expect(tags[0]?.name).toBe('client')
    expect(tags[1]?.name).toBe('vip')
  })

  it('should find or create a tag', () => {
    const repo = createTagRepo(db)

    const tag1 = repo.findOrCreate('vip')
    expect(tag1.name).toBe('vip')
    expect(tag1.id).toBe(1)

    const tag2 = repo.findOrCreate('vip')
    expect(tag2.id).toBe(1)
  })

  it('should apply a tag to a person', () => {
    const repo = createTagRepo(db)

    const tagging = repo.apply('person', 1, 'vip')
    expect(tagging.entity_type).toBe('person')
    expect(tagging.entity_id).toBe(1)
  })

  it('should not duplicate taggings', () => {
    const repo = createTagRepo(db)

    repo.apply('person', 1, 'vip')
    repo.apply('person', 1, 'vip')

    const tags = repo.getForEntity('person', 1)
    expect(tags).toHaveLength(1)
  })

  it('should remove a tag from an entity', () => {
    const repo = createTagRepo(db)

    repo.apply('person', 1, 'vip')
    const removed = repo.remove('person', 1, 'vip')

    expect(removed).toBe(true)
    expect(repo.getForEntity('person', 1)).toHaveLength(0)
  })

  it('should return false when removing nonexistent tag', () => {
    const repo = createTagRepo(db)
    expect(repo.remove('person', 1, 'nonexistent')).toBe(false)
  })

  it('should get tags for an entity', () => {
    const repo = createTagRepo(db)

    repo.apply('person', 1, 'vip')
    repo.apply('person', 1, 'client')

    const tags = repo.getForEntity('person', 1)
    expect(tags).toHaveLength(2)
    expect(tags.map((t) => t.name).sort()).toEqual(['client', 'vip'])
  })

  it('should get entity IDs for a tag', () => {
    const repo = createTagRepo(db)

    repo.apply('person', 1, 'vip')
    repo.apply('person', 2, 'vip')

    const ids = repo.getEntities('person', 'vip')
    expect(ids).toHaveLength(2)
    expect(ids.sort()).toEqual([1, 2])
  })

  it('should return empty array for nonexistent tag', () => {
    const repo = createTagRepo(db)
    expect(repo.getEntities('person', 'nonexistent')).toEqual([])
  })

  it('should apply tags to different entity types', () => {
    const repo = createTagRepo(db)

    repo.apply('person', 1, 'important')
    repo.apply('organization', 1, 'important')

    expect(repo.getForEntity('person', 1)).toHaveLength(1)
    expect(repo.getForEntity('organization', 1)).toHaveLength(1)
    expect(repo.getEntities('person', 'important')).toEqual([1])
    expect(repo.getEntities('organization', 'important')).toEqual([1])
  })

  it('should delete a tag and all its taggings', () => {
    const repo = createTagRepo(db)

    repo.apply('person', 1, 'temp')
    repo.apply('person', 2, 'temp')

    const deleted = repo.delete('temp')
    expect(deleted).toBe(true)

    expect(repo.findAll()).toHaveLength(0)
    expect(repo.getForEntity('person', 1)).toHaveLength(0)
    expect(repo.getForEntity('person', 2)).toHaveLength(0)
  })

  it('should return false when deleting nonexistent tag', () => {
    const repo = createTagRepo(db)
    expect(repo.delete('nonexistent')).toBe(false)
  })
})
