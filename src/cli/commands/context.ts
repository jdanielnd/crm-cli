import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createCustomFieldRepo } from '../../db/repositories/custom-field.repo.js'
import { createDealRepo } from '../../db/repositories/deal.repo.js'
import { createInteractionRepo } from '../../db/repositories/interaction.repo.js'
import { createOrgRepo } from '../../db/repositories/org.repo.js'
import { createPersonRepo } from '../../db/repositories/person.repo.js'
import { createRelationshipRepo } from '../../db/repositories/relationship.repo.js'
import { createTagRepo } from '../../db/repositories/tag.repo.js'
import { createTaskRepo } from '../../db/repositories/task.repo.js'
import { resolveFormat } from '../../formatters/index.js'
import { NotFoundError } from '../../models/errors.js'

interface GlobalOpts {
  format?: string
  quiet?: boolean
  db?: string
  color?: boolean
}

function getOpts(cmd: Command): GlobalOpts {
  let root = cmd
  while (root.parent) root = root.parent
  return root.opts()
}

export function registerContextCommands(program: Command): void {
  program
    .command('context')
    .description('Full context briefing for a person')
    .argument('<id>', 'Person ID')
    .option('--limit <n>', 'Max interactions to show', '10')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const id = Number(idStr)
      const limit = Number(opts['limit'])

      const personRepo = createPersonRepo(db)
      const orgRepo = createOrgRepo(db)
      const interactionRepo = createInteractionRepo(db)
      const dealRepo = createDealRepo(db)
      const taskRepo = createTaskRepo(db)
      const tagRepo = createTagRepo(db)
      const relRepo = createRelationshipRepo(db)
      const cfRepo = createCustomFieldRepo(db)

      const person = personRepo.findById(id)
      if (!person) throw new NotFoundError('person', id)

      const org = person.org_id ? orgRepo.findById(person.org_id) : undefined
      const interactions = interactionRepo.findAll({ personId: id, limit })
      const deals = dealRepo.findAll({ personId: id })
      const tasks = taskRepo.findAll({ personId: id, completed: false })
      const tags = tagRepo.getForEntity('person', id)
      const relationships = relRepo.findForPerson(id)
      const customFields = cfRepo.get('person', id)

      const format = resolveFormat(globalOpts.format)

      if (format === 'json') {
        const data = {
          person,
          organization: org ?? null,
          tags: tags.map((t) => t.name),
          custom_fields: customFields,
          relationships: relationships.map((r) => ({
            id: r.id,
            person_id: r.person_id === id ? r.related_person_id : r.person_id,
            type: r.type,
            notes: r.notes,
          })),
          recent_interactions: interactions.map((i) => ({ ...i })),
          open_deals: deals.filter((d) => d.stage !== 'won' && d.stage !== 'lost'),
          open_tasks: tasks,
          stats: {
            total_interactions: interactions.length,
            total_deals: deals.length,
            open_tasks: tasks.length,
          },
        }
        process.stdout.write(JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0) + '\n')
      } else {
        const name = [person.first_name, person.last_name].filter(Boolean).join(' ')
        const lines: string[] = []

        lines.push(`=== ${name} ===`)
        if (person.email) lines.push(`Email: ${person.email}`)
        if (person.phone) lines.push(`Phone: ${person.phone}`)
        if (person.title) lines.push(`Title: ${person.title}`)
        if (person.company) lines.push(`Company: ${person.company}`)
        if (org) lines.push(`Organization: ${org.name}`)
        if (person.location) lines.push(`Location: ${person.location}`)
        if (tags.length > 0) lines.push(`Tags: ${tags.map((t) => t.name).join(', ')}`)
        if (person.summary) lines.push(`\nSummary: ${person.summary}`)
        if (person.notes) lines.push(`Notes: ${person.notes}`)

        if (customFields.length > 0) {
          lines.push('\n--- Custom Fields ---')
          for (const cf of customFields) {
            lines.push(`${cf.field_name}: ${cf.field_value ?? ''}`)
          }
        }

        if (relationships.length > 0) {
          lines.push('\n--- Relationships ---')
          for (const r of relationships) {
            const otherId = r.person_id === id ? r.related_person_id : r.person_id
            const other = personRepo.findById(otherId)
            const otherName = other
              ? [other.first_name, other.last_name].filter(Boolean).join(' ')
              : `#${String(otherId)}`
            lines.push(`${r.type}: ${otherName}`)
          }
        }

        if (interactions.length > 0) {
          lines.push('\n--- Recent Interactions ---')
          for (const i of interactions) {
            const subject = i.subject ?? '(no subject)'
            lines.push(`${i.occurred_at} [${i.type}] ${subject}`)
          }
        }

        const openDeals = deals.filter((d) => d.stage !== 'won' && d.stage !== 'lost')
        if (openDeals.length > 0) {
          lines.push('\n--- Open Deals ---')
          for (const d of openDeals) {
            const value = d.value ? ` ($${String(d.value)})` : ''
            lines.push(`#${String(d.id)} ${d.title} [${d.stage}]${value}`)
          }
        }

        if (tasks.length > 0) {
          lines.push('\n--- Open Tasks ---')
          for (const t of tasks) {
            const due = t.due_at ? ` due ${t.due_at}` : ''
            lines.push(`#${String(t.id)} ${t.title} [${t.priority}]${due}`)
          }
        }

        process.stdout.write(lines.join('\n') + '\n')
      }
    })
}
