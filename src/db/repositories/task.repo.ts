import type Database from 'better-sqlite3'

import type { Priority, TaskInsert, TaskRow, TaskUpdate } from '../../models/types.js'

export interface TaskFilters {
  personId?: number
  dealId?: number
  priority?: Priority
  overdue?: boolean
  completed?: boolean
  sort?: 'due' | 'priority' | 'created'
  limit?: number
  offset?: number
}

export function createTaskRepo(db: Database.Database) {
  return {
    create(input: TaskInsert): TaskRow {
      const stmt = db.prepare(`
        INSERT INTO tasks (title, description, due_at, priority, person_id, deal_id)
        VALUES (@title, @description, @due_at, @priority, @person_id, @deal_id)
      `)

      const result = stmt.run({
        title: input.title,
        description: input.description ?? null,
        due_at: input.due_at ?? null,
        priority: input.priority ?? 'normal',
        person_id: input.person_id ?? null,
        deal_id: input.deal_id ?? null,
      })

      return db.prepare('SELECT * FROM tasks WHERE id = ?').get(result.lastInsertRowid) as TaskRow
    },

    findById(id: number): TaskRow | undefined {
      return db.prepare('SELECT * FROM tasks WHERE id = ? AND archived = 0').get(id) as
        | TaskRow
        | undefined
    },

    findAll(filters?: TaskFilters): TaskRow[] {
      const conditions = ['t.archived = 0']
      const params: Record<string, unknown> = {}

      if (filters?.personId !== undefined) {
        conditions.push('t.person_id = @personId')
        params['personId'] = filters.personId
      }

      if (filters?.dealId !== undefined) {
        conditions.push('t.deal_id = @dealId')
        params['dealId'] = filters.dealId
      }

      if (filters?.priority) {
        conditions.push('t.priority = @priority')
        params['priority'] = filters.priority
      }

      if (filters?.overdue) {
        conditions.push("t.due_at < datetime('now') AND t.completed = 0")
      }

      if (filters?.completed !== undefined) {
        conditions.push(`t.completed = @completed`)
        params['completed'] = filters.completed ? 1 : 0
      }

      let orderBy = 't.id DESC'
      if (filters?.sort === 'due') orderBy = 't.due_at ASC NULLS LAST'
      else if (filters?.sort === 'priority')
        orderBy =
          "CASE t.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 WHEN 'low' THEN 4 END"
      else if (filters?.sort === 'created') orderBy = 't.created_at DESC'

      const limit = filters?.limit ?? 100
      const offset = filters?.offset ?? 0

      const sql = `
        SELECT t.* FROM tasks t
        WHERE ${conditions.join(' AND ')}
        ORDER BY ${orderBy}
        LIMIT @limit OFFSET @offset
      `

      return db.prepare(sql).all({ ...params, limit, offset }) as TaskRow[]
    },

    update(id: number, input: TaskUpdate): TaskRow | undefined {
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

      db.prepare(`UPDATE tasks SET ${fields.join(', ')} WHERE id = @id AND archived = 0`).run(
        params,
      )

      return this.findById(id)
    },

    complete(id: number): TaskRow | undefined {
      db.prepare(
        "UPDATE tasks SET completed = 1, completed_at = datetime('now'), updated_at = datetime('now') WHERE id = ? AND archived = 0",
      ).run(id)
      return this.findById(id)
    },

    archive(id: number): boolean {
      const result = db
        .prepare(
          "UPDATE tasks SET archived = 1, updated_at = datetime('now') WHERE id = ? AND archived = 0",
        )
        .run(id)
      return result.changes > 0
    },
  }
}
