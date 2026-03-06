import type Database from 'better-sqlite3'

import type { PersonInsert, PersonRow, PersonUpdate } from '../../models/types.js'

export interface PersonFilters {
  tag?: string
  orgId?: number
  search?: string
  sort?: 'name' | 'created' | 'updated'
  limit?: number
  offset?: number
}

export function createPersonRepo(db: Database.Database) {
  return {
    create(input: PersonInsert): PersonRow {
      const stmt = db.prepare(`
        INSERT INTO people (first_name, last_name, email, phone, title, company, location, linkedin, twitter, website, notes, summary, org_id)
        VALUES (@first_name, @last_name, @email, @phone, @title, @company, @location, @linkedin, @twitter, @website, @notes, @summary, @org_id)
      `)

      const result = stmt.run({
        first_name: input.first_name,
        last_name: input.last_name ?? null,
        email: input.email ?? null,
        phone: input.phone ?? null,
        title: input.title ?? null,
        company: input.company ?? null,
        location: input.location ?? null,
        linkedin: input.linkedin ?? null,
        twitter: input.twitter ?? null,
        website: input.website ?? null,
        notes: input.notes ?? null,
        summary: input.summary ?? null,
        org_id: input.org_id ?? null,
      })

      return db
        .prepare('SELECT * FROM people WHERE id = ?')
        .get(result.lastInsertRowid) as PersonRow
    },

    findById(id: number): PersonRow | undefined {
      return db.prepare('SELECT * FROM people WHERE id = ? AND archived = 0').get(id) as
        | PersonRow
        | undefined
    },

    findAll(filters?: PersonFilters): PersonRow[] {
      const conditions = ['p.archived = 0']
      const params: Record<string, unknown> = {}

      if (filters?.orgId !== undefined) {
        conditions.push('p.org_id = @orgId')
        params['orgId'] = filters.orgId
      }

      if (filters?.tag) {
        conditions.push(`p.id IN (
          SELECT tg.entity_id FROM taggings tg
          JOIN tags t ON t.id = tg.tag_id
          WHERE tg.entity_type = 'person' AND t.name = @tag
        )`)
        params['tag'] = filters.tag
      }

      if (filters?.search) {
        conditions.push('p.id IN (SELECT rowid FROM people_fts WHERE people_fts MATCH @search)')
        params['search'] = filters.search
      }

      let orderBy = 'p.id DESC'
      if (filters?.sort === 'name') orderBy = 'p.first_name ASC, p.last_name ASC'
      else if (filters?.sort === 'created') orderBy = 'p.created_at DESC'
      else if (filters?.sort === 'updated') orderBy = 'p.updated_at DESC'

      const limit = filters?.limit ?? 100
      const offset = filters?.offset ?? 0

      const sql = `
        SELECT p.* FROM people p
        WHERE ${conditions.join(' AND ')}
        ORDER BY ${orderBy}
        LIMIT @limit OFFSET @offset
      `

      return db.prepare(sql).all({ ...params, limit, offset }) as PersonRow[]
    },

    update(id: number, input: PersonUpdate): PersonRow | undefined {
      const fields: string[] = []
      const params: Record<string, unknown> = { id }

      for (const [key, value] of Object.entries(input)) {
        if (value !== undefined) {
          fields.push(`${key} = @${key}`)
          params[key] = value ?? null
        }
      }

      if (fields.length === 0) {
        return this.findById(id)
      }

      fields.push("updated_at = datetime('now')")

      db.prepare(`UPDATE people SET ${fields.join(', ')} WHERE id = @id AND archived = 0`).run(
        params,
      )

      return this.findById(id)
    },

    archive(id: number): boolean {
      const result = db
        .prepare(
          "UPDATE people SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0",
        )
        .run(id)
      return result.changes > 0
    },

    search(query: string, limit = 20): PersonRow[] {
      return db
        .prepare(
          `SELECT p.* FROM people p
           JOIN people_fts fts ON fts.rowid = p.id
           WHERE people_fts MATCH ? AND p.archived = 0
           ORDER BY rank
           LIMIT ?`,
        )
        .all(query, limit) as PersonRow[]
    },
  }
}
