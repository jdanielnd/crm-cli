import type Database from 'better-sqlite3'

import type { OrganizationInsert, OrganizationRow, OrganizationUpdate } from '../../models/types.js'

export interface OrgFilters {
  search?: string
  sort?: 'name' | 'created' | 'updated'
  limit?: number
  offset?: number
}

export function createOrgRepo(db: Database.Database) {
  return {
    create(input: OrganizationInsert): OrganizationRow {
      const stmt = db.prepare(`
        INSERT INTO organizations (name, domain, industry, location, notes, summary)
        VALUES (@name, @domain, @industry, @location, @notes, @summary)
      `)

      const result = stmt.run({
        name: input.name,
        domain: input.domain ?? null,
        industry: input.industry ?? null,
        location: input.location ?? null,
        notes: input.notes ?? null,
        summary: input.summary ?? null,
      })

      return db
        .prepare('SELECT * FROM organizations WHERE id = ?')
        .get(result.lastInsertRowid) as OrganizationRow
    },

    findById(id: number): OrganizationRow | undefined {
      return db.prepare('SELECT * FROM organizations WHERE id = ? AND archived = 0').get(id) as
        | OrganizationRow
        | undefined
    },

    findAll(filters?: OrgFilters): OrganizationRow[] {
      const conditions = ['o.archived = 0']
      const params: Record<string, unknown> = {}

      if (filters?.search) {
        conditions.push(
          'o.id IN (SELECT rowid FROM organizations_fts WHERE organizations_fts MATCH @search)',
        )
        params['search'] = filters.search
      }

      let orderBy = 'o.id DESC'
      if (filters?.sort === 'name') orderBy = 'o.name ASC'
      else if (filters?.sort === 'created') orderBy = 'o.created_at DESC'
      else if (filters?.sort === 'updated') orderBy = 'o.updated_at DESC'

      const limit = filters?.limit ?? 100
      const offset = filters?.offset ?? 0

      const sql = `
        SELECT o.* FROM organizations o
        WHERE ${conditions.join(' AND ')}
        ORDER BY ${orderBy}
        LIMIT @limit OFFSET @offset
      `

      return db.prepare(sql).all({ ...params, limit, offset }) as OrganizationRow[]
    },

    update(id: number, input: OrganizationUpdate): OrganizationRow | undefined {
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

      db.prepare(
        `UPDATE organizations SET ${fields.join(', ')} WHERE id = @id AND archived = 0`,
      ).run(params)

      return this.findById(id)
    },

    archive(id: number): boolean {
      const result = db
        .prepare(
          "UPDATE organizations SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0",
        )
        .run(id)
      return result.changes > 0
    },

    search(query: string, limit = 20): OrganizationRow[] {
      return db
        .prepare(
          `SELECT o.* FROM organizations o
           JOIN organizations_fts fts ON fts.rowid = o.id
           WHERE organizations_fts MATCH ? AND o.archived = 0
           ORDER BY rank
           LIMIT ?`,
        )
        .all(query, limit) as OrganizationRow[]
    },
  }
}
