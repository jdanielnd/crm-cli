import type Database from 'better-sqlite3'

import type { DealInsert, DealRow, DealStage, DealUpdate } from '../../models/types.js'

export interface DealFilters {
  stage?: DealStage | DealStage[]
  personId?: number
  orgId?: number
  sort?: 'value' | 'created' | 'updated' | 'stage'
  limit?: number
  offset?: number
}

export interface PipelineSummary {
  stage: string
  count: number
  total_value: number
}

export function createDealRepo(db: Database.Database) {
  return {
    create(input: DealInsert): DealRow {
      const stmt = db.prepare(`
        INSERT INTO deals (title, value, currency, stage, person_id, org_id, closed_at, notes)
        VALUES (@title, @value, @currency, @stage, @person_id, @org_id, @closed_at, @notes)
      `)

      const result = stmt.run({
        title: input.title,
        value: input.value ?? null,
        currency: input.currency ?? 'USD',
        stage: input.stage ?? 'lead',
        person_id: input.person_id ?? null,
        org_id: input.org_id ?? null,
        closed_at: input.closed_at ?? null,
        notes: input.notes ?? null,
      })

      return db.prepare('SELECT * FROM deals WHERE id = ?').get(result.lastInsertRowid) as DealRow
    },

    findById(id: number): DealRow | undefined {
      return db.prepare('SELECT * FROM deals WHERE id = ? AND archived = 0').get(id) as
        | DealRow
        | undefined
    },

    findAll(filters?: DealFilters): DealRow[] {
      const conditions = ['d.archived = 0']
      const params: Record<string, unknown> = {}

      if (filters?.stage) {
        if (Array.isArray(filters.stage)) {
          const placeholders = filters.stage.map((_, i) => `@stage${String(i)}`)
          conditions.push(`d.stage IN (${placeholders.join(', ')})`)
          for (let i = 0; i < filters.stage.length; i++) {
            params[`stage${String(i)}`] = filters.stage[i]
          }
        } else {
          conditions.push('d.stage = @stage')
          params['stage'] = filters.stage
        }
      }

      if (filters?.personId !== undefined) {
        conditions.push('d.person_id = @personId')
        params['personId'] = filters.personId
      }

      if (filters?.orgId !== undefined) {
        conditions.push('d.org_id = @orgId')
        params['orgId'] = filters.orgId
      }

      let orderBy = 'd.id DESC'
      if (filters?.sort === 'value') orderBy = 'd.value DESC NULLS LAST'
      else if (filters?.sort === 'created') orderBy = 'd.created_at DESC'
      else if (filters?.sort === 'updated') orderBy = 'd.updated_at DESC'
      else if (filters?.sort === 'stage') orderBy = 'd.stage ASC'

      const limit = filters?.limit ?? 100
      const offset = filters?.offset ?? 0

      const sql = `
        SELECT d.* FROM deals d
        WHERE ${conditions.join(' AND ')}
        ORDER BY ${orderBy}
        LIMIT @limit OFFSET @offset
      `

      return db.prepare(sql).all({ ...params, limit, offset }) as DealRow[]
    },

    update(id: number, input: DealUpdate): DealRow | undefined {
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

      db.prepare(`UPDATE deals SET ${fields.join(', ')} WHERE id = @id AND archived = 0`).run(
        params,
      )

      return this.findById(id)
    },

    archive(id: number): boolean {
      const result = db
        .prepare(
          "UPDATE deals SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0",
        )
        .run(id)
      return result.changes > 0
    },

    pipeline(): PipelineSummary[] {
      return db
        .prepare(
          `SELECT stage, COUNT(*) as count, COALESCE(SUM(value), 0) as total_value
           FROM deals
           WHERE archived = 0 AND stage NOT IN ('won', 'lost')
           GROUP BY stage
           ORDER BY CASE stage
             WHEN 'lead' THEN 1
             WHEN 'prospect' THEN 2
             WHEN 'proposal' THEN 3
             WHEN 'negotiation' THEN 4
           END`,
        )
        .all() as PipelineSummary[]
    },

    search(query: string, limit = 20): DealRow[] {
      return db
        .prepare(
          `SELECT d.* FROM deals d
           JOIN deals_fts fts ON fts.rowid = d.id
           WHERE deals_fts MATCH ? AND d.archived = 0
           ORDER BY rank
           LIMIT ?`,
        )
        .all(query, limit) as DealRow[]
    },
  }
}
