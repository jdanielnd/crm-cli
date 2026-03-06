import type Database from 'better-sqlite3'

import type { TaggableEntity, TaggingRow, TagRow } from '../../models/types.js'

export function createTagRepo(db: Database.Database) {
  return {
    findAll(): TagRow[] {
      return db.prepare('SELECT * FROM tags ORDER BY name ASC').all() as TagRow[]
    },

    findOrCreate(name: string): TagRow {
      const existing = db.prepare('SELECT * FROM tags WHERE name = ?').get(name) as
        | TagRow
        | undefined
      if (existing) return existing

      const result = db.prepare('INSERT INTO tags (name) VALUES (?)').run(name)
      return db.prepare('SELECT * FROM tags WHERE id = ?').get(result.lastInsertRowid) as TagRow
    },

    apply(entityType: TaggableEntity, entityId: number, tagName: string): TaggingRow {
      const tag = this.findOrCreate(tagName)

      db.prepare(
        `INSERT OR IGNORE INTO taggings (tag_id, entity_type, entity_id) VALUES (?, ?, ?)`,
      ).run(tag.id, entityType, entityId)

      return db
        .prepare('SELECT * FROM taggings WHERE tag_id = ? AND entity_type = ? AND entity_id = ?')
        .get(tag.id, entityType, entityId) as TaggingRow
    },

    remove(entityType: TaggableEntity, entityId: number, tagName: string): boolean {
      const tag = db.prepare('SELECT * FROM tags WHERE name = ?').get(tagName) as TagRow | undefined
      if (!tag) return false

      const result = db
        .prepare('DELETE FROM taggings WHERE tag_id = ? AND entity_type = ? AND entity_id = ?')
        .run(tag.id, entityType, entityId)
      return result.changes > 0
    },

    getForEntity(entityType: TaggableEntity, entityId: number): TagRow[] {
      return db
        .prepare(
          `SELECT t.* FROM tags t
           JOIN taggings tg ON tg.tag_id = t.id
           WHERE tg.entity_type = ? AND tg.entity_id = ?
           ORDER BY t.name ASC`,
        )
        .all(entityType, entityId) as TagRow[]
    },

    getEntities(entityType: TaggableEntity, tagName: string): number[] {
      const tag = db.prepare('SELECT * FROM tags WHERE name = ?').get(tagName) as TagRow | undefined
      if (!tag) return []

      return db
        .prepare('SELECT entity_id FROM taggings WHERE tag_id = ? AND entity_type = ?')
        .all(tag.id, entityType)
        .map((r) => (r as { entity_id: number }).entity_id)
    },

    delete(tagName: string): boolean {
      const tag = db.prepare('SELECT * FROM tags WHERE name = ?').get(tagName) as TagRow | undefined
      if (!tag) return false

      db.prepare('DELETE FROM taggings WHERE tag_id = ?').run(tag.id)
      db.prepare('DELETE FROM tags WHERE id = ?').run(tag.id)
      return true
    },
  }
}
