import Database from 'better-sqlite3'
import { describe, expect, it, beforeEach } from 'vitest'

import { runMigrations } from '../../src/db/index.js'
import { createTaskRepo } from '../../src/db/repositories/task.repo.js'

let db: Database.Database

function setup() {
  db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

describe('TaskRepository', () => {
  beforeEach(() => {
    db = setup()
  })

  it('should create a task', () => {
    const repo = createTaskRepo(db)
    const task = repo.create({ title: 'Follow up with Jane', priority: 'high' })

    expect(task.id).toBe(1)
    expect(task.title).toBe('Follow up with Jane')
    expect(task.priority).toBe('high')
    expect(task.completed).toBe(0)
    expect(task.uuid).toBeTruthy()
    expect(task.archived).toBe(0)
  })

  it('should create a task with minimal fields', () => {
    const repo = createTaskRepo(db)
    const task = repo.create({ title: 'Quick task' })

    expect(task.title).toBe('Quick task')
    expect(task.priority).toBe('normal')
    expect(task.due_at).toBeNull()
    expect(task.person_id).toBeNull()
  })

  it('should create a task linked to a person', () => {
    const repo = createTaskRepo(db)
    db.prepare("INSERT INTO people (first_name) VALUES ('Jane')").run()

    const task = repo.create({ title: 'Call Jane', person_id: 1 })
    expect(task.person_id).toBe(1)
  })

  it('should find a task by ID', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Test' })

    const task = repo.findById(1)
    expect(task).toBeTruthy()
    expect(task?.title).toBe('Test')
  })

  it('should return undefined for missing task', () => {
    const repo = createTaskRepo(db)
    expect(repo.findById(999)).toBeUndefined()
  })

  it('should list all tasks', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'A' })
    repo.create({ title: 'B' })

    const tasks = repo.findAll()
    expect(tasks).toHaveLength(2)
  })

  it('should filter by person ID', () => {
    const repo = createTaskRepo(db)
    db.prepare("INSERT INTO people (first_name) VALUES ('Jane')").run()

    repo.create({ title: 'Jane task', person_id: 1 })
    repo.create({ title: 'Other task' })

    const results = repo.findAll({ personId: 1 })
    expect(results).toHaveLength(1)
    expect(results[0]?.title).toBe('Jane task')
  })

  it('should filter by priority', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Urgent', priority: 'urgent' })
    repo.create({ title: 'Normal', priority: 'normal' })

    const results = repo.findAll({ priority: 'urgent' })
    expect(results).toHaveLength(1)
    expect(results[0]?.title).toBe('Urgent')
  })

  it('should filter completed tasks', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Done' })
    repo.create({ title: 'Pending' })
    repo.complete(1)

    const completed = repo.findAll({ completed: true })
    expect(completed).toHaveLength(1)
    expect(completed[0]?.title).toBe('Done')

    const pending = repo.findAll({ completed: false })
    expect(pending).toHaveLength(1)
    expect(pending[0]?.title).toBe('Pending')
  })

  it('should sort by priority', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Low', priority: 'low' })
    repo.create({ title: 'Urgent', priority: 'urgent' })
    repo.create({ title: 'Normal', priority: 'normal' })

    const results = repo.findAll({ sort: 'priority' })
    expect(results[0]?.title).toBe('Urgent')
    expect(results[1]?.title).toBe('Normal')
    expect(results[2]?.title).toBe('Low')
  })

  it('should limit results', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'A' })
    repo.create({ title: 'B' })
    repo.create({ title: 'C' })

    const results = repo.findAll({ limit: 2 })
    expect(results).toHaveLength(2)
  })

  it('should update a task', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Old title' })

    const updated = repo.update(1, { title: 'New title', priority: 'high' })
    expect(updated?.title).toBe('New title')
    expect(updated?.priority).toBe('high')
  })

  it('should return undefined when updating nonexistent task', () => {
    const repo = createTaskRepo(db)
    expect(repo.update(999, { title: 'Test' })).toBeUndefined()
  })

  it('should complete a task', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Do it' })

    const task = repo.complete(1)
    expect(task?.completed).toBe(1)
    expect(task?.completed_at).toBeTruthy()
  })

  it('should archive a task', () => {
    const repo = createTaskRepo(db)
    repo.create({ title: 'Test' })

    expect(repo.archive(1)).toBe(true)
    expect(repo.findAll()).toHaveLength(0)
    expect(repo.findById(1)).toBeUndefined()
  })

  it('should return false when archiving nonexistent task', () => {
    const repo = createTaskRepo(db)
    expect(repo.archive(999)).toBe(false)
  })
})
