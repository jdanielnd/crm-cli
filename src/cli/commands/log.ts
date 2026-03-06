import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import {
  createInteractionRepo,
  type InteractionFilters,
} from '../../db/repositories/interaction.repo.js'
import {
  type ColumnDef,
  type FormatOptions,
  formatOutput,
  resolveFormat,
} from '../../formatters/index.js'
import { NotFoundError, ValidationError } from '../../models/errors.js'
import { INTERACTION_TYPES, type InteractionType } from '../../models/types.js'

const INTERACTION_COLUMNS: ColumnDef[] = [
  { key: 'id', header: 'ID' },
  { key: 'type', header: 'Type' },
  { key: 'subject', header: 'Subject' },
  { key: 'direction', header: 'Dir' },
  { key: 'occurred_at', header: 'Date' },
  { key: 'person_ids', header: 'People' },
]

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

function buildFormatOpts(globalOpts: GlobalOpts, columns?: ColumnDef[]): FormatOptions {
  const opts: FormatOptions = {}
  if (globalOpts.quiet !== undefined) opts.quiet = globalOpts.quiet
  if (globalOpts.color === false) opts.noColor = true
  if (columns) opts.columns = columns
  return opts
}

export function registerLogCommands(program: Command): void {
  const log = program.command('log').description('Log interactions')

  for (const type of INTERACTION_TYPES) {
    log
      .command(type)
      .description(`Log a ${type}`)
      .argument('<person-ids...>', 'Person ID(s)')
      .option('--subject <text>', 'Subject line')
      .option('--content <text>', 'Content / body')
      .option('--direction <dir>', 'Direction: inbound, outbound')
      .option('--at <datetime>', 'When it occurred (default: now)')
      .action((personIdStrs: string[], opts: Record<string, unknown>, cmd: Command) => {
        const globalOpts = getOpts(cmd)
        const db = getDb(globalOpts.db)
        const repo = createInteractionRepo(db)

        const personIds = personIdStrs.map((s) => {
          const n = Number(s)
          if (Number.isNaN(n) || n <= 0) throw new ValidationError(`Invalid person ID: "${s}"`)
          return n
        })

        const direction = opts['direction'] as string | undefined
        if (direction && direction !== 'inbound' && direction !== 'outbound') {
          throw new ValidationError(`Invalid direction: "${direction}". Use inbound or outbound`)
        }

        const interaction = repo.create(
          {
            type,
            subject: (opts['subject'] as string | undefined) ?? null,
            content: (opts['content'] as string | undefined) ?? null,
            direction: (direction as 'inbound' | 'outbound' | undefined) ?? null,
            occurred_at: opts['at'] as string | undefined,
          },
          personIds,
        )

        const format = resolveFormat(globalOpts.format)
        process.stdout.write(
          formatOutput(
            [{ ...interaction }],
            format,
            buildFormatOpts(globalOpts, INTERACTION_COLUMNS),
          ) + '\n',
        )
        if (!globalOpts.quiet) {
          process.stderr.write(
            `Logged ${type} #${String(interaction.id)} with person(s) ${personIds.join(', ')}\n`,
          )
        }
      })
  }

  // List interactions
  log
    .command('list')
    .description('List interactions')
    .option('--person <id>', 'Filter by person ID')
    .option('--org <id>', 'Filter by organization ID')
    .option('--type <type>', `Filter by type: ${INTERACTION_TYPES.join(', ')}`)
    .option('--since <date>', 'Filter interactions after date')
    .option('--sort <field>', 'Sort by: occurred, created')
    .option('--limit <n>', 'Max results', '50')
    .action((opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createInteractionRepo(db)

      const filters: InteractionFilters = { limit: Number(opts['limit']) }
      if (opts['person']) filters.personId = Number(opts['person'])
      if (opts['org']) filters.orgId = Number(opts['org'])
      if (opts['type']) filters.type = opts['type'] as InteractionType
      if (opts['since']) filters.since = opts['since'] as string
      if (opts['sort']) filters.sort = opts['sort'] as 'occurred' | 'created'

      const interactions = repo.findAll(filters)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput(
          interactions.map((i) => ({ ...i })),
          format,
          buildFormatOpts(globalOpts, INTERACTION_COLUMNS),
        ) + '\n',
      )
    })

  // Show interaction detail
  log
    .command('show')
    .description('Show interaction details')
    .argument('<id>', 'Interaction ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createInteractionRepo(db)
      const id = Number(idStr)

      const interaction = repo.findById(id)
      if (!interaction) throw new NotFoundError('interaction', id)

      const format = resolveFormat(globalOpts.format)
      if (format === 'json') {
        process.stdout.write(JSON.stringify(interaction, null, process.stdout.isTTY ? 2 : 0) + '\n')
      } else {
        process.stdout.write(
          formatOutput(
            [{ ...interaction }],
            format,
            buildFormatOpts(globalOpts, INTERACTION_COLUMNS),
          ) + '\n',
        )
      }
    })

  // Delete interaction
  log
    .command('delete')
    .description('Delete an interaction (soft-delete)')
    .argument('<id>', 'Interaction ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createInteractionRepo(db)
      const id = Number(idStr)

      const archived = repo.archive(id)
      if (!archived) throw new NotFoundError('interaction', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Archived interaction #${String(id)}\n`)
      }
    })
}
