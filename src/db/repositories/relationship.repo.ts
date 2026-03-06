import type Database from 'better-sqlite3'

import type { RelationshipInsert, RelationshipRow, RelationshipType } from '../../models/types.js'

export function createRelationshipRepo(db: Database.Database) {
  return {
    create(input: RelationshipInsert): RelationshipRow {
      const stmt = db.prepare(`
        INSERT INTO relationships (person_id, related_person_id, type, notes)
        VALUES (@person_id, @related_person_id, @type, @notes)
      `)

      const result = stmt.run({
        person_id: input.person_id,
        related_person_id: input.related_person_id,
        type: input.type,
        notes: input.notes ?? null,
      })

      return db
        .prepare('SELECT * FROM relationships WHERE id = ?')
        .get(result.lastInsertRowid) as RelationshipRow
    },

    findForPerson(personId: number): RelationshipRow[] {
      return db
        .prepare(
          'SELECT * FROM relationships WHERE person_id = ? OR related_person_id = ? ORDER BY type ASC',
        )
        .all(personId, personId) as RelationshipRow[]
    },

    findByType(personId: number, type: RelationshipType): RelationshipRow[] {
      return db
        .prepare(
          'SELECT * FROM relationships WHERE (person_id = ? OR related_person_id = ?) AND type = ?',
        )
        .all(personId, personId, type) as RelationshipRow[]
    },

    delete(id: number): boolean {
      const result = db.prepare('DELETE FROM relationships WHERE id = ?').run(id)
      return result.changes > 0
    },
  }
}
