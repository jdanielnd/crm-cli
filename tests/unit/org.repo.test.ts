import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createOrgRepo } from '../../src/db/repositories/org.repo.js'
import { createCustomFieldRepo } from '../../src/db/repositories/custom-field.repo.js'

let db: Database.Database

function setup() {
  db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

describe('OrgRepository', () => {
  beforeEach(() => {
    db = setup()
  })

  it('should create an organization', () => {
    const repo = createOrgRepo(db)
    const org = repo.create({ name: 'Acme Corp', domain: 'acme.com', industry: 'SaaS' })

    expect(org.id).toBe(1)
    expect(org.name).toBe('Acme Corp')
    expect(org.domain).toBe('acme.com')
    expect(org.industry).toBe('SaaS')
    expect(org.uuid).toBeTruthy()
    expect(org.archived).toBe(0)
  })

  it('should create an organization with minimal fields', () => {
    const repo = createOrgRepo(db)
    const org = repo.create({ name: 'Acme' })

    expect(org.id).toBe(1)
    expect(org.name).toBe('Acme')
    expect(org.domain).toBeNull()
    expect(org.industry).toBeNull()
  })

  it('should find an organization by ID', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme' })

    const org = repo.findById(1)
    expect(org).toBeTruthy()
    expect(org?.name).toBe('Acme')
  })

  it('should return undefined for missing organization', () => {
    const repo = createOrgRepo(db)
    expect(repo.findById(999)).toBeUndefined()
  })

  it('should list all organizations', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme' })
    repo.create({ name: 'Globex' })

    const orgs = repo.findAll()
    expect(orgs).toHaveLength(2)
  })

  it('should sort by name', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Zebra Inc' })
    repo.create({ name: 'Alpha Corp' })

    const orgs = repo.findAll({ sort: 'name' })
    expect(orgs[0]?.name).toBe('Alpha Corp')
    expect(orgs[1]?.name).toBe('Zebra Inc')
  })

  it('should limit results', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'A' })
    repo.create({ name: 'B' })
    repo.create({ name: 'C' })

    const orgs = repo.findAll({ limit: 2 })
    expect(orgs).toHaveLength(2)
  })

  it('should update an organization', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme', domain: 'old.com' })

    const updated = repo.update(1, { domain: 'new.com' })
    expect(updated?.domain).toBe('new.com')
    expect(updated?.name).toBe('Acme')
  })

  it('should return undefined when updating nonexistent organization', () => {
    const repo = createOrgRepo(db)
    expect(repo.update(999, { name: 'Test' })).toBeUndefined()
  })

  it('should archive (soft delete) an organization', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme' })

    const archived = repo.archive(1)
    expect(archived).toBe(true)

    expect(repo.findAll()).toHaveLength(0)
    expect(repo.findById(1)).toBeUndefined()
  })

  it('should return false when archiving nonexistent organization', () => {
    const repo = createOrgRepo(db)
    expect(repo.archive(999)).toBe(false)
  })

  it('should search via FTS', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme Corp', notes: 'Enterprise SaaS company' })
    repo.create({ name: 'Globex', notes: 'Hardware manufacturer' })

    const results = repo.search('SaaS')
    expect(results).toHaveLength(1)
    expect(results[0]?.name).toBe('Acme Corp')
  })

  it('should not find archived organizations in search', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme Corp', notes: 'archived org' })
    repo.archive(1)

    const results = repo.search('archived')
    expect(results).toHaveLength(0)
  })

  it('should filter by search in findAll', () => {
    const repo = createOrgRepo(db)
    repo.create({ name: 'Acme Corp' })
    repo.create({ name: 'Globex' })

    const results = repo.findAll({ search: 'Acme' })
    expect(results).toHaveLength(1)
    expect(results[0]?.name).toBe('Acme Corp')
  })
})

describe('CustomFieldRepository with organizations', () => {
  beforeEach(() => {
    db = setup()
  })

  it('should set and get custom fields for organizations', () => {
    const cfRepo = createCustomFieldRepo(db)
    db.prepare("INSERT INTO organizations (name) VALUES ('Acme')").run()

    cfRepo.set('organization', 1, 'founded', '2010')
    cfRepo.set('organization', 1, 'employees', '500')

    const fields = cfRepo.get('organization', 1)
    expect(fields).toHaveLength(2)
    expect(fields.find((f) => f.field_name === 'founded')?.field_value).toBe('2010')
    expect(fields.find((f) => f.field_name === 'employees')?.field_value).toBe('500')
  })
})
