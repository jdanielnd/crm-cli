import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createDealRepo } from '../../db/repositories/deal.repo.js'
import { createInteractionRepo } from '../../db/repositories/interaction.repo.js'
import { createOrgRepo } from '../../db/repositories/org.repo.js'
import { createPersonRepo } from '../../db/repositories/person.repo.js'
import { type FormatOptions, formatOutput, resolveFormat } from '../../formatters/index.js'

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

function buildFormatOpts(globalOpts: GlobalOpts): FormatOptions {
  const opts: FormatOptions = {}
  if (globalOpts.quiet !== undefined) opts.quiet = globalOpts.quiet
  if (globalOpts.color === false) opts.noColor = true
  return opts
}

type SearchResult = {
  type: string
  id: number
  title: string
  detail: string
}

export function registerSearchCommands(program: Command): void {
  program
    .command('search')
    .description('Search across all entities')
    .argument('<query>', 'Search query')
    .option('--type <type>', 'Filter by type: person, org, interaction, deal')
    .option('--limit <n>', 'Max results per type', '10')
    .action((query: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const limit = Number(opts['limit'])
      const typeFilter = opts['type'] as string | undefined

      const results: SearchResult[] = []

      if (!typeFilter || typeFilter === 'person') {
        const personRepo = createPersonRepo(db)
        const people = personRepo.search(query, limit)
        for (const p of people) {
          results.push({
            type: 'person',
            id: p.id,
            title: [p.first_name, p.last_name].filter(Boolean).join(' '),
            detail: p.email ?? p.company ?? '',
          })
        }
      }

      if (!typeFilter || typeFilter === 'org') {
        const orgRepo = createOrgRepo(db)
        const orgs = orgRepo.search(query, limit)
        for (const o of orgs) {
          results.push({
            type: 'org',
            id: o.id,
            title: o.name,
            detail: o.domain ?? o.industry ?? '',
          })
        }
      }

      if (!typeFilter || typeFilter === 'interaction') {
        const interactionRepo = createInteractionRepo(db)
        const interactions = interactionRepo.search(query, limit)
        for (const i of interactions) {
          results.push({
            type: 'interaction',
            id: i.id,
            title: i.subject ?? `${i.type} (no subject)`,
            detail: i.occurred_at,
          })
        }
      }

      if (!typeFilter || typeFilter === 'deal') {
        const dealRepo = createDealRepo(db)
        const deals = dealRepo.search(query, limit)
        for (const d of deals) {
          results.push({
            type: 'deal',
            id: d.id,
            title: d.title,
            detail: `${d.stage}${d.value ? ` $${String(d.value)}` : ''}`,
          })
        }
      }

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(formatOutput(results, format, buildFormatOpts(globalOpts)) + '\n')
    })
}
