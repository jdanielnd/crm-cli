import type Database from 'better-sqlite3'

import type { CustomFieldRow } from '../../models/types.js'

export function createCustomFieldRepo(db: Database.Database) {
  return {
    set(entityType: string, entityId: number, fieldName: string, fieldValue: string | null): void {
      db.prepare(
        `INSERT INTO custom_fields (entity_type, entity_id, field_name, field_value)
         VALUES (?, ?, ?, ?)
         ON CONFLICT (entity_type, entity_id, field_name)
         DO UPDATE SET field_value = excluded.field_value`,
      ).run(entityType, entityId, fieldName, fieldValue)
    },

    get(entityType: string, entityId: number): CustomFieldRow[] {
      return db
        .prepare('SELECT * FROM custom_fields WHERE entity_type = ? AND entity_id = ?')
        .all(entityType, entityId) as CustomFieldRow[]
    },

    delete(entityType: string, entityId: number, fieldName: string): boolean {
      const result = db
        .prepare(
          'DELETE FROM custom_fields WHERE entity_type = ? AND entity_id = ? AND field_name = ?',
        )
        .run(entityType, entityId, fieldName)
      return result.changes > 0
    },

    deleteAll(entityType: string, entityId: number): void {
      db.prepare('DELETE FROM custom_fields WHERE entity_type = ? AND entity_id = ?').run(
        entityType,
        entityId,
      )
    },
  }
}
