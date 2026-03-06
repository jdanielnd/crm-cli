import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createDealRepo } from '../../src/db/repositories/deal.repo.js'

let db: Database.Database

function setup() {
  db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

describe('DealRepository', () => {
  beforeEach(() => {
    db = setup()
  })

  it('should create a deal', () => {
    const repo = createDealRepo(db)
    const deal = repo.create({ title: 'Website Redesign', value: 15000, stage: 'proposal' })

    expect(deal.id).toBe(1)
    expect(deal.title).toBe('Website Redesign')
    expect(deal.value).toBe(15000)
    expect(deal.stage).toBe('proposal')
    expect(deal.currency).toBe('USD')
    expect(deal.uuid).toBeTruthy()
    expect(deal.archived).toBe(0)
  })

  it('should create a deal with minimal fields', () => {
    const repo = createDealRepo(db)
    const deal = repo.create({ title: 'Quick Deal' })

    expect(deal.title).toBe('Quick Deal')
    expect(deal.stage).toBe('lead')
    expect(deal.value).toBeNull()
  })

  it('should find a deal by ID', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Test Deal' })

    const deal = repo.findById(1)
    expect(deal).toBeTruthy()
    expect(deal?.title).toBe('Test Deal')
  })

  it('should return undefined for missing deal', () => {
    const repo = createDealRepo(db)
    expect(repo.findById(999)).toBeUndefined()
  })

  it('should list all deals', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Deal A' })
    repo.create({ title: 'Deal B' })

    const deals = repo.findAll()
    expect(deals).toHaveLength(2)
  })

  it('should filter by stage', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Lead Deal', stage: 'lead' })
    repo.create({ title: 'Proposal Deal', stage: 'proposal' })

    const results = repo.findAll({ stage: 'proposal' })
    expect(results).toHaveLength(1)
    expect(results[0]?.title).toBe('Proposal Deal')
  })

  it('should filter by multiple stages', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Lead Deal', stage: 'lead' })
    repo.create({ title: 'Proposal Deal', stage: 'proposal' })
    repo.create({ title: 'Won Deal', stage: 'won' })

    const results = repo.findAll({ stage: ['lead', 'proposal'] })
    expect(results).toHaveLength(2)
  })

  it('should filter by person ID', () => {
    const repo = createDealRepo(db)
    db.prepare("INSERT INTO people (first_name) VALUES ('Jane')").run()
    repo.create({ title: 'Jane Deal', person_id: 1 })
    repo.create({ title: 'Other Deal' })

    const results = repo.findAll({ personId: 1 })
    expect(results).toHaveLength(1)
    expect(results[0]?.title).toBe('Jane Deal')
  })

  it('should sort by value', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Small', value: 1000 })
    repo.create({ title: 'Big', value: 50000 })

    const results = repo.findAll({ sort: 'value' })
    expect(results[0]?.title).toBe('Big')
  })

  it('should limit results', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'A' })
    repo.create({ title: 'B' })
    repo.create({ title: 'C' })

    const results = repo.findAll({ limit: 2 })
    expect(results).toHaveLength(2)
  })

  it('should update a deal', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Test', stage: 'lead' })

    const updated = repo.update(1, { stage: 'won', closed_at: '2026-03-06' })
    expect(updated?.stage).toBe('won')
    expect(updated?.closed_at).toBe('2026-03-06')
  })

  it('should return undefined when updating nonexistent deal', () => {
    const repo = createDealRepo(db)
    expect(repo.update(999, { title: 'Test' })).toBeUndefined()
  })

  it('should archive a deal', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Test' })

    expect(repo.archive(1)).toBe(true)
    expect(repo.findAll()).toHaveLength(0)
    expect(repo.findById(1)).toBeUndefined()
  })

  it('should return false when archiving nonexistent deal', () => {
    const repo = createDealRepo(db)
    expect(repo.archive(999)).toBe(false)
  })

  it('should generate pipeline summary', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'A', value: 10000, stage: 'lead' })
    repo.create({ title: 'B', value: 20000, stage: 'lead' })
    repo.create({ title: 'C', value: 50000, stage: 'proposal' })
    repo.create({ title: 'D', value: 5000, stage: 'won' })

    const pipeline = repo.pipeline()
    expect(pipeline).toHaveLength(2)

    const lead = pipeline.find((p) => p.stage === 'lead')
    expect(lead?.count).toBe(2)
    expect(lead?.total_value).toBe(30000)

    const proposal = pipeline.find((p) => p.stage === 'proposal')
    expect(proposal?.count).toBe(1)
    expect(proposal?.total_value).toBe(50000)
  })

  it('should search via FTS', () => {
    const repo = createDealRepo(db)
    repo.create({ title: 'Website Redesign', notes: 'Complete overhaul' })
    repo.create({ title: 'Mobile App' })

    const results = repo.search('redesign')
    expect(results).toHaveLength(1)
    expect(results[0]?.title).toBe('Website Redesign')
  })
})
