import Database from 'better-sqlite3'
import { describe, expect, it } from 'vitest'

import { runMigrations } from '../../src/db/index.js'

function createTestDb(): Database.Database {
  const db = new Database(':memory:')
  db.pragma('foreign_keys = ON')
  runMigrations(db)
  return db
}

describe('Migration runner', () => {
  it('should apply migrations and set user_version', () => {
    const db = createTestDb()
    const version = db.pragma('user_version', { simple: true }) as number
    expect(version).toBe(1)
    db.close()
  })

  it('should be idempotent (running twice does nothing)', () => {
    const db = createTestDb()
    runMigrations(db)
    const version = db.pragma('user_version', { simple: true }) as number
    expect(version).toBe(1)
    db.close()
  })
})

describe('Schema', () => {
  it('should create all expected tables', () => {
    const db = createTestDb()
    const tables = db
      .prepare(
        "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name",
      )
      .all() as { name: string }[]

    const tableNames = tables.map((t) => t.name)

    expect(tableNames).toContain('people')
    expect(tableNames).toContain('organizations')
    expect(tableNames).toContain('interactions')
    expect(tableNames).toContain('interaction_people')
    expect(tableNames).toContain('deals')
    expect(tableNames).toContain('tasks')
    expect(tableNames).toContain('tags')
    expect(tableNames).toContain('taggings')
    expect(tableNames).toContain('custom_fields')
    expect(tableNames).toContain('relationships')

    db.close()
  })

  it('should create FTS5 virtual tables', () => {
    const db = createTestDb()
    const tables = db
      .prepare(
        "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%_fts' ORDER BY name",
      )
      .all() as { name: string }[]

    const tableNames = tables.map((t) => t.name)

    expect(tableNames).toContain('people_fts')
    expect(tableNames).toContain('organizations_fts')
    expect(tableNames).toContain('interactions_fts')
    expect(tableNames).toContain('deals_fts')

    db.close()
  })

  it('should insert and retrieve a person', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name, last_name, email) VALUES (?, ?, ?)').run(
      'Jane',
      'Smith',
      'jane@example.com',
    )

    const person = db.prepare('SELECT * FROM people WHERE id = 1').get() as Record<string, unknown>

    expect(person['first_name']).toBe('Jane')
    expect(person['last_name']).toBe('Smith')
    expect(person['email']).toBe('jane@example.com')
    expect(person['uuid']).toBeTruthy()
    expect(person['archived']).toBe(0)

    db.close()
  })

  it('should auto-populate uuid on insert', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Alice')
    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Bob')

    const alice = db.prepare('SELECT uuid FROM people WHERE id = 1').get() as { uuid: string }
    const bob = db.prepare('SELECT uuid FROM people WHERE id = 2').get() as { uuid: string }

    expect(alice.uuid).toBeTruthy()
    expect(bob.uuid).toBeTruthy()
    expect(alice.uuid).not.toBe(bob.uuid)

    db.close()
  })

  it('should enforce foreign key constraints', () => {
    const db = createTestDb()

    expect(() => {
      db.prepare('INSERT INTO people (first_name, org_id) VALUES (?, ?)').run('Jane', 999)
    }).toThrow()

    db.close()
  })

  it('should sync FTS on insert', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name, last_name, email) VALUES (?, ?, ?)').run(
      'Jane',
      'Smith',
      'jane@example.com',
    )

    const results = db
      .prepare("SELECT rowid FROM people_fts WHERE people_fts MATCH 'Jane'")
      .all() as { rowid: number }[]

    expect(results).toHaveLength(1)
    expect(results[0]?.rowid).toBe(1)

    db.close()
  })

  it('should sync FTS on update', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare('UPDATE people SET first_name = ? WHERE id = 1').run('Janet')

    const oldResults = db
      .prepare("SELECT rowid FROM people_fts WHERE people_fts MATCH 'Jane'")
      .all()
    const newResults = db
      .prepare("SELECT rowid FROM people_fts WHERE people_fts MATCH 'Janet'")
      .all()

    expect(oldResults).toHaveLength(0)
    expect(newResults).toHaveLength(1)

    db.close()
  })

  it('should support organization-person relationship', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO organizations (name) VALUES (?)').run('Acme Corp')
    db.prepare('INSERT INTO people (first_name, org_id) VALUES (?, ?)').run('Jane', 1)

    const person = db.prepare('SELECT org_id FROM people WHERE id = 1').get() as {
      org_id: number
    }
    expect(person.org_id).toBe(1)

    db.close()
  })

  it('should support interaction with multiple people', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Bob')
    db.prepare("INSERT INTO interactions (type, subject) VALUES ('meeting', 'Demo')").run()
    db.prepare('INSERT INTO interaction_people (interaction_id, person_id) VALUES (?, ?)').run(1, 1)
    db.prepare('INSERT INTO interaction_people (interaction_id, person_id) VALUES (?, ?)').run(1, 2)

    const linked = db
      .prepare('SELECT person_id FROM interaction_people WHERE interaction_id = 1')
      .all() as { person_id: number }[]
    expect(linked).toHaveLength(2)

    db.close()
  })

  it('should support polymorphic tagging', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare("INSERT INTO tags (name) VALUES ('vip')").run()
    db.prepare(
      "INSERT INTO taggings (tag_id, entity_type, entity_id) VALUES (1, 'person', 1)",
    ).run()

    const tagging = db
      .prepare('SELECT * FROM taggings WHERE entity_type = ? AND entity_id = ?')
      .get('person', 1) as Record<string, unknown>
    expect(tagging['tag_id']).toBe(1)

    db.close()
  })

  it('should enforce unique constraint on taggings', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare("INSERT INTO tags (name) VALUES ('vip')").run()
    db.prepare(
      "INSERT INTO taggings (tag_id, entity_type, entity_id) VALUES (1, 'person', 1)",
    ).run()

    expect(() => {
      db.prepare(
        "INSERT INTO taggings (tag_id, entity_type, entity_id) VALUES (1, 'person', 1)",
      ).run()
    }).toThrow()

    db.close()
  })

  it('should support deal stages', () => {
    const db = createTestDb()

    db.prepare(
      "INSERT INTO deals (title, stage, value) VALUES ('Big Deal', 'proposal', 50000)",
    ).run()

    const deal = db.prepare('SELECT * FROM deals WHERE id = 1').get() as Record<string, unknown>
    expect(deal['stage']).toBe('proposal')
    expect(deal['value']).toBe(50000)

    db.close()
  })

  it('should reject invalid deal stage', () => {
    const db = createTestDb()

    expect(() => {
      db.prepare("INSERT INTO deals (title, stage) VALUES ('Bad Deal', 'invalid')").run()
    }).toThrow()

    db.close()
  })

  it('should support tasks with priority', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare(
      "INSERT INTO tasks (title, priority, person_id, due_at) VALUES ('Follow up', 'high', 1, '2026-03-15')",
    ).run()

    const task = db.prepare('SELECT * FROM tasks WHERE id = 1').get() as Record<string, unknown>
    expect(task['priority']).toBe('high')
    expect(task['person_id']).toBe(1)
    expect(task['completed']).toBe(0)

    db.close()
  })

  it('should support custom fields', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare(
      "INSERT INTO custom_fields (entity_type, entity_id, field_name, field_value) VALUES ('person', 1, 'birthday', '1990-03-15')",
    ).run()

    const field = db
      .prepare(
        "SELECT field_value FROM custom_fields WHERE entity_type = 'person' AND entity_id = 1 AND field_name = 'birthday'",
      )
      .get() as { field_value: string }
    expect(field.field_value).toBe('1990-03-15')

    db.close()
  })

  it('should support person-to-person relationships', () => {
    const db = createTestDb()

    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Jane')
    db.prepare('INSERT INTO people (first_name) VALUES (?)').run('Bob')
    db.prepare(
      "INSERT INTO relationships (person_id, related_person_id, type) VALUES (1, 2, 'colleague')",
    ).run()

    const rel = db.prepare('SELECT * FROM relationships WHERE person_id = 1').get() as Record<
      string,
      unknown
    >
    expect(rel['related_person_id']).toBe(2)
    expect(rel['type']).toBe('colleague')

    db.close()
  })
})
