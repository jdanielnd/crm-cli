import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { z } from 'zod'

import { getDb } from '../db/index.js'
import { createCustomFieldRepo } from '../db/repositories/custom-field.repo.js'
import { createDealRepo } from '../db/repositories/deal.repo.js'
import { createInteractionRepo } from '../db/repositories/interaction.repo.js'
import { createOrgRepo } from '../db/repositories/org.repo.js'
import { createPersonRepo } from '../db/repositories/person.repo.js'
import { createRelationshipRepo } from '../db/repositories/relationship.repo.js'
import { createTagRepo } from '../db/repositories/tag.repo.js'
import { createTaskRepo } from '../db/repositories/task.repo.js'
import { DEAL_STAGES, INTERACTION_TYPES, PRIORITIES, RELATIONSHIP_TYPES } from '../models/types.js'

function text(s: string) {
  return { content: [{ type: 'text' as const, text: s }] }
}

function json(data: unknown) {
  return text(JSON.stringify(data, null, 2))
}

function notFound(entity: string, id: number) {
  return {
    content: [{ type: 'text' as const, text: `${entity} #${String(id)} not found` }],
    isError: true as const,
  }
}

export function createMcpServer(dbPath?: string) {
  const server = new McpServer({ name: 'crm', version: '0.1.0' })
  const db = getDb(dbPath)

  server.registerTool(
    'crm_person_search',
    {
      description: 'Search people by name, email, tag, or org',
      inputSchema: {
        query: z.string().optional(),
        tag: z.string().optional(),
        org_id: z.number().optional(),
        limit: z.number().optional(),
      },
    },
    ({ query, tag, org_id, limit }) => {
      const repo = createPersonRepo(db)
      if (query) return json(repo.search(query, limit ?? 20))
      const filters: Record<string, unknown> = {}
      if (tag) filters['tag'] = tag
      if (org_id) filters['orgId'] = org_id
      if (limit) filters['limit'] = limit
      return json(repo.findAll(filters))
    },
  )

  server.registerTool(
    'crm_person_get',
    {
      description: 'Get full details for a person including custom fields, tags, and relationships',
      inputSchema: { id: z.number() },
    },
    ({ id }) => {
      const person = createPersonRepo(db).findById(id)
      if (!person) return notFound('Person', id)
      return json({
        ...person,
        custom_fields: createCustomFieldRepo(db).get('person', id),
        tags: createTagRepo(db)
          .getForEntity('person', id)
          .map((t) => t.name),
        relationships: createRelationshipRepo(db).findForPerson(id),
      })
    },
  )

  server.registerTool(
    'crm_person_create',
    {
      description: 'Create a new person',
      inputSchema: {
        first_name: z.string(),
        last_name: z.string().optional(),
        email: z.string().optional(),
        phone: z.string().optional(),
        title: z.string().optional(),
        company: z.string().optional(),
        location: z.string().optional(),
        notes: z.string().optional(),
        org_id: z.number().optional(),
      },
    },
    (input) => {
      return json(
        createPersonRepo(db).create({
          first_name: input.first_name,
          last_name: input.last_name ?? null,
          email: input.email ?? null,
          phone: input.phone ?? null,
          title: input.title ?? null,
          company: input.company ?? null,
          location: input.location ?? null,
          notes: input.notes ?? null,
          org_id: input.org_id ?? null,
        }),
      )
    },
  )

  server.registerTool(
    'crm_person_update',
    {
      description: 'Update person fields',
      inputSchema: {
        id: z.number(),
        first_name: z.string().optional(),
        last_name: z.string().optional(),
        email: z.string().optional(),
        phone: z.string().optional(),
        title: z.string().optional(),
        company: z.string().optional(),
        location: z.string().optional(),
        notes: z.string().optional(),
        org_id: z.number().optional(),
      },
    },
    ({ id, ...fields }) => {
      const person = createPersonRepo(db).update(id, fields)
      if (!person) return notFound('Person', id)
      return json(person)
    },
  )

  server.registerTool(
    'crm_person_update_summary',
    {
      description: 'Update the AI-maintained summary/dossier for a person',
      inputSchema: { id: z.number(), summary: z.string() },
    },
    ({ id, summary }) => {
      const person = createPersonRepo(db).update(id, { summary })
      if (!person) return notFound('Person', id)
      return text(`Updated summary for person #${String(id)}`)
    },
  )

  server.registerTool(
    'crm_org_search',
    {
      description: 'Search organizations',
      inputSchema: { query: z.string().optional(), limit: z.number().optional() },
    },
    ({ query, limit }) => {
      const repo = createOrgRepo(db)
      return json(query ? repo.search(query, limit ?? 20) : repo.findAll({ limit: limit ?? 100 }))
    },
  )

  server.registerTool(
    'crm_org_get',
    {
      description: 'Get org details with people',
      inputSchema: { id: z.number() },
    },
    ({ id }) => {
      const org = createOrgRepo(db).findById(id)
      if (!org) return notFound('Organization', id)
      return json({ ...org, people: createPersonRepo(db).findAll({ orgId: id }) })
    },
  )

  server.registerTool(
    'crm_interaction_log',
    {
      description: 'Log an interaction (call, email, meeting, note, message)',
      inputSchema: {
        type: z.enum(INTERACTION_TYPES),
        person_ids: z.array(z.number()),
        subject: z.string().optional(),
        content: z.string().optional(),
        direction: z.enum(['inbound', 'outbound']).optional(),
        occurred_at: z.string().optional(),
      },
    },
    (input) => {
      const interaction = createInteractionRepo(db).create(
        {
          type: input.type,
          subject: input.subject ?? null,
          content: input.content ?? null,
          direction: input.direction ?? null,
          occurred_at: input.occurred_at,
        },
        input.person_ids,
      )
      return json({ ...interaction })
    },
  )

  server.registerTool(
    'crm_interaction_list',
    {
      description: 'List interactions for a person or org',
      inputSchema: {
        person_id: z.number().optional(),
        org_id: z.number().optional(),
        type: z.enum(INTERACTION_TYPES).optional(),
        limit: z.number().optional(),
      },
    },
    (input) => {
      const filters: Record<string, unknown> = {}
      if (input.person_id) filters['personId'] = input.person_id
      if (input.org_id) filters['orgId'] = input.org_id
      if (input.type) filters['type'] = input.type
      if (input.limit) filters['limit'] = input.limit
      return json(
        createInteractionRepo(db)
          .findAll(filters)
          .map((r) => ({ ...r })),
      )
    },
  )

  server.registerTool(
    'crm_search',
    {
      description: 'Cross-entity full-text search',
      inputSchema: {
        query: z.string(),
        type: z.enum(['person', 'org', 'interaction', 'deal']).optional(),
        limit: z.number().optional(),
      },
    },
    ({ query, type: typeFilter, limit }) => {
      const results: Array<{ type: string; id: number; title: string; detail: string }> = []
      const max = limit ?? 10
      if (!typeFilter || typeFilter === 'person') {
        for (const p of createPersonRepo(db).search(query, max)) {
          results.push({
            type: 'person',
            id: p.id,
            title: [p.first_name, p.last_name].filter(Boolean).join(' '),
            detail: p.email ?? '',
          })
        }
      }
      if (!typeFilter || typeFilter === 'org') {
        for (const o of createOrgRepo(db).search(query, max)) {
          results.push({ type: 'org', id: o.id, title: o.name, detail: o.domain ?? '' })
        }
      }
      if (!typeFilter || typeFilter === 'interaction') {
        for (const i of createInteractionRepo(db).search(query, max)) {
          results.push({
            type: 'interaction',
            id: i.id,
            title: i.subject ?? i.type,
            detail: i.occurred_at,
          })
        }
      }
      if (!typeFilter || typeFilter === 'deal') {
        for (const d of createDealRepo(db).search(query, max)) {
          results.push({ type: 'deal', id: d.id, title: d.title, detail: d.stage })
        }
      }
      return json(results)
    },
  )

  server.registerTool(
    'crm_context',
    {
      description: 'Full context for a person (the briefing)',
      inputSchema: { id: z.number(), interaction_limit: z.number().optional() },
    },
    ({ id, interaction_limit }) => {
      const person = createPersonRepo(db).findById(id)
      if (!person) return notFound('Person', id)
      const org = person.org_id ? createOrgRepo(db).findById(person.org_id) : undefined
      return json({
        person,
        organization: org ?? null,
        tags: createTagRepo(db)
          .getForEntity('person', id)
          .map((t) => t.name),
        custom_fields: createCustomFieldRepo(db).get('person', id),
        relationships: createRelationshipRepo(db).findForPerson(id),
        recent_interactions: createInteractionRepo(db)
          .findAll({ personId: id, limit: interaction_limit ?? 10 })
          .map((i) => ({ ...i })),
        open_deals: createDealRepo(db)
          .findAll({ personId: id })
          .filter((d) => d.stage !== 'won' && d.stage !== 'lost'),
        open_tasks: createTaskRepo(db).findAll({ personId: id, completed: false }),
      })
    },
  )

  server.registerTool(
    'crm_task_create',
    {
      description: 'Create a follow-up task',
      inputSchema: {
        title: z.string(),
        person_id: z.number().optional(),
        deal_id: z.number().optional(),
        due_at: z.string().optional(),
        priority: z.enum(PRIORITIES).optional(),
        description: z.string().optional(),
      },
    },
    (input) => {
      return json(
        createTaskRepo(db).create({
          title: input.title,
          person_id: input.person_id ?? null,
          deal_id: input.deal_id ?? null,
          due_at: input.due_at ?? null,
          priority: input.priority,
          description: input.description ?? null,
        }),
      )
    },
  )

  server.registerTool(
    'crm_task_list',
    {
      description: 'List open tasks',
      inputSchema: {
        person_id: z.number().optional(),
        deal_id: z.number().optional(),
        overdue: z.boolean().optional(),
        limit: z.number().optional(),
      },
    },
    (input) => {
      const filters: Record<string, unknown> = { completed: false }
      if (input.person_id) filters['personId'] = input.person_id
      if (input.deal_id) filters['dealId'] = input.deal_id
      if (input.overdue) filters['overdue'] = true
      if (input.limit) filters['limit'] = input.limit
      return json(createTaskRepo(db).findAll(filters))
    },
  )

  server.registerTool(
    'crm_deal_create',
    {
      description: 'Create a deal',
      inputSchema: {
        title: z.string(),
        value: z.number().optional(),
        stage: z.enum(DEAL_STAGES).optional(),
        person_id: z.number().optional(),
        org_id: z.number().optional(),
        notes: z.string().optional(),
      },
    },
    (input) => {
      return json(
        createDealRepo(db).create({
          title: input.title,
          value: input.value ?? null,
          stage: input.stage,
          person_id: input.person_id ?? null,
          org_id: input.org_id ?? null,
          notes: input.notes ?? null,
        }),
      )
    },
  )

  server.registerTool(
    'crm_deal_update',
    {
      description: 'Update deal stage or fields',
      inputSchema: {
        id: z.number(),
        title: z.string().optional(),
        value: z.number().optional(),
        stage: z.enum(DEAL_STAGES).optional(),
        closed_at: z.string().optional(),
        notes: z.string().optional(),
      },
    },
    ({ id, ...fields }) => {
      const deal = createDealRepo(db).update(id, fields)
      if (!deal) return notFound('Deal', id)
      return json(deal)
    },
  )

  server.registerTool(
    'crm_tag_apply',
    {
      description: 'Apply tag to an entity',
      inputSchema: {
        entity_type: z.enum(['person', 'organization', 'deal', 'interaction']),
        entity_id: z.number(),
        tag: z.string(),
      },
    },
    ({ entity_type, entity_id, tag }) => {
      createTagRepo(db).apply(entity_type, entity_id, tag)
      return text(`Tagged ${entity_type} #${String(entity_id)} as "${tag}"`)
    },
  )

  server.registerTool(
    'crm_person_relate',
    {
      description: 'Create a relationship between two people',
      inputSchema: {
        person_id: z.number(),
        related_person_id: z.number(),
        type: z.enum(RELATIONSHIP_TYPES),
        notes: z.string().optional(),
      },
    },
    (input) => {
      return json(
        createRelationshipRepo(db).create({
          person_id: input.person_id,
          related_person_id: input.related_person_id,
          type: input.type,
          notes: input.notes ?? null,
        }),
      )
    },
  )

  server.registerTool(
    'crm_stats',
    {
      description: 'CRM summary stats',
    },
    () => {
      const people = (
        db.prepare('SELECT COUNT(*) as count FROM people WHERE archived = 0').get() as {
          count: number
        }
      ).count
      const orgs = (
        db.prepare('SELECT COUNT(*) as count FROM organizations WHERE archived = 0').get() as {
          count: number
        }
      ).count
      const openDeals = db
        .prepare(
          "SELECT COUNT(*) as count, COALESCE(SUM(value), 0) as total_value FROM deals WHERE archived = 0 AND stage NOT IN ('won', 'lost')",
        )
        .get() as { count: number; total_value: number }
      const overdueTasks = (
        db
          .prepare(
            "SELECT COUNT(*) as count FROM tasks WHERE archived = 0 AND completed = 0 AND due_at < datetime('now')",
          )
          .get() as { count: number }
      ).count
      const interactionsThisWeek = (
        db
          .prepare(
            "SELECT COUNT(*) as count FROM interactions WHERE archived = 0 AND occurred_at >= datetime('now', '-7 days')",
          )
          .get() as { count: number }
      ).count
      return json({
        people,
        organizations: orgs,
        open_deals: openDeals.count,
        open_deals_value: openDeals.total_value,
        overdue_tasks: overdueTasks,
        interactions_this_week: interactionsThisWeek,
      })
    },
  )

  return server
}

export async function startMcpServer(dbPath?: string) {
  const server = createMcpServer(dbPath)
  const transport = new StdioServerTransport()
  await server.connect(transport)

  const shutdown = () => {
    void server.close()
  }
  process.on('SIGINT', shutdown)
  process.on('SIGTERM', shutdown)
}
