import type Database from 'better-sqlite3'

import type { InteractionInsert, InteractionRow, InteractionType } from '../../models/types.js'

export interface InteractionWithPeople extends InteractionRow {
  person_ids: number[]
}

export interface InteractionFilters {
  personId?: number
  orgId?: number
  type?: InteractionType
  since?: string
  sort?: 'occurred' | 'created'
  limit?: number
  offset?: number
}

export function createInteractionRepo(db: Database.Database) {
  return {
    create(input: InteractionInsert, personIds: number[]): InteractionWithPeople {
      const stmt = db.prepare(`
        INSERT INTO interactions (type, subject, content, direction, occurred_at)
        VALUES (@type, @subject, @content, @direction, COALESCE(@occurred_at, datetime('now')))
      `)

      const result = stmt.run({
        type: input.type,
        subject: input.subject ?? null,
        content: input.content ?? null,
        direction: input.direction ?? null,
        occurred_at: input.occurred_at ?? null,
      })

      const id = result.lastInsertRowid as number

      if (personIds.length > 0) {
        const linkStmt = db.prepare(
          'INSERT INTO interaction_people (interaction_id, person_id) VALUES (?, ?)',
        )
        for (const pid of personIds) {
          linkStmt.run(id, pid)
        }
      }

      const row = db.prepare('SELECT * FROM interactions WHERE id = ?').get(id) as InteractionRow
      return { ...row, person_ids: personIds }
    },

    findById(id: number): InteractionWithPeople | undefined {
      const row = db.prepare('SELECT * FROM interactions WHERE id = ? AND archived = 0').get(id) as
        | InteractionRow
        | undefined
      if (!row) return undefined

      const personIds = db
        .prepare('SELECT person_id FROM interaction_people WHERE interaction_id = ?')
        .all(id)
        .map((r) => (r as { person_id: number }).person_id)

      return { ...row, person_ids: personIds }
    },

    findAll(filters?: InteractionFilters): InteractionWithPeople[] {
      const conditions = ['i.archived = 0']
      const params: Record<string, unknown> = {}

      if (filters?.personId !== undefined) {
        conditions.push(
          'i.id IN (SELECT interaction_id FROM interaction_people WHERE person_id = @personId)',
        )
        params['personId'] = filters.personId
      }

      if (filters?.orgId !== undefined) {
        conditions.push(`i.id IN (
          SELECT ip.interaction_id FROM interaction_people ip
          JOIN people p ON p.id = ip.person_id
          WHERE p.org_id = @orgId
        )`)
        params['orgId'] = filters.orgId
      }

      if (filters?.type) {
        conditions.push('i.type = @type')
        params['type'] = filters.type
      }

      if (filters?.since) {
        conditions.push('i.occurred_at >= @since')
        params['since'] = filters.since
      }

      let orderBy = 'i.occurred_at DESC'
      if (filters?.sort === 'created') orderBy = 'i.created_at DESC'

      const limit = filters?.limit ?? 100
      const offset = filters?.offset ?? 0

      const sql = `
        SELECT i.* FROM interactions i
        WHERE ${conditions.join(' AND ')}
        ORDER BY ${orderBy}
        LIMIT @limit OFFSET @offset
      `

      const rows = db.prepare(sql).all({ ...params, limit, offset }) as InteractionRow[]

      return rows.map((row) => {
        const personIds = db
          .prepare('SELECT person_id FROM interaction_people WHERE interaction_id = ?')
          .all(row.id)
          .map((r) => (r as { person_id: number }).person_id)
        return { ...row, person_ids: personIds }
      })
    },

    archive(id: number): boolean {
      const result = db
        .prepare(
          "UPDATE interactions SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0",
        )
        .run(id)
      return result.changes > 0
    },

    search(query: string, limit = 20): InteractionWithPeople[] {
      const rows = db
        .prepare(
          `SELECT i.* FROM interactions i
           JOIN interactions_fts fts ON fts.rowid = i.id
           WHERE interactions_fts MATCH ? AND i.archived = 0
           ORDER BY rank
           LIMIT ?`,
        )
        .all(query, limit) as InteractionRow[]

      return rows.map((row) => {
        const personIds = db
          .prepare('SELECT person_id FROM interaction_people WHERE interaction_id = ?')
          .all(row.id)
          .map((r) => (r as { person_id: number }).person_id)
        return { ...row, person_ids: personIds }
      })
    },
  }
}
