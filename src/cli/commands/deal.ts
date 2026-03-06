import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createCustomFieldRepo } from '../../db/repositories/custom-field.repo.js'
import { createDealRepo, type DealFilters } from '../../db/repositories/deal.repo.js'
import {
  type ColumnDef,
  type FormatOptions,
  formatOutput,
  resolveFormat,
} from '../../formatters/index.js'
import { NotFoundError } from '../../models/errors.js'
import { DEAL_STAGES, type DealStage } from '../../models/types.js'

const DEAL_COLUMNS: ColumnDef[] = [
  { key: 'id', header: 'ID' },
  { key: 'title', header: 'Title' },
  { key: 'value', header: 'Value' },
  { key: 'stage', header: 'Stage' },
  { key: 'person_id', header: 'Person' },
  { key: 'org_id', header: 'Org' },
]

const PIPELINE_COLUMNS: ColumnDef[] = [
  { key: 'stage', header: 'Stage' },
  { key: 'count', header: 'Deals' },
  { key: 'total_value', header: 'Total Value' },
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

export function registerDealCommands(program: Command): void {
  const deal = program.command('deal').description('Manage deals')

  deal
    .command('add')
    .description('Add a new deal')
    .argument('<title>', 'Deal title')
    .option('--value <amount>', 'Deal value')
    .option('--currency <code>', 'Currency code (default: USD)')
    .option('--stage <stage>', `Stage: ${DEAL_STAGES.join(', ')}`, 'lead')
    .option('--person <id>', 'Person ID')
    .option('--org <id>', 'Organization ID')
    .option('--note <text>', 'Notes')
    .action((title: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createDealRepo(db)

      const deal = repo.create({
        title,
        value: opts['value'] ? Number(opts['value']) : null,
        currency: opts['currency'] as string | undefined,
        stage: opts['stage'] as DealStage | undefined,
        person_id: opts['person'] ? Number(opts['person']) : null,
        org_id: opts['org'] ? Number(opts['org']) : null,
        notes: (opts['note'] as string | undefined) ?? null,
      })

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([deal], format, buildFormatOpts(globalOpts, DEAL_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Created deal #${String(deal.id)}\n`)
      }
    })

  deal
    .command('list')
    .description('List deals')
    .option('--stage <stages>', 'Filter by stage(s), comma-separated')
    .option('--person <id>', 'Filter by person ID')
    .option('--org <id>', 'Filter by organization ID')
    .option('--sort <field>', 'Sort by: value, created, updated, stage')
    .option('--limit <n>', 'Max results', '100')
    .action((opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createDealRepo(db)

      const filters: DealFilters = { limit: Number(opts['limit']) }
      if (opts['stage']) {
        const stages = (opts['stage'] as string).split(',') as DealStage[]
        if (stages.length === 1 && stages[0]) {
          filters.stage = stages[0]
        } else {
          filters.stage = stages
        }
      }
      if (opts['person']) filters.personId = Number(opts['person'])
      if (opts['org']) filters.orgId = Number(opts['org'])
      if (opts['sort']) filters.sort = opts['sort'] as 'value' | 'created' | 'updated' | 'stage'

      const deals = repo.findAll(filters)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput(deals, format, buildFormatOpts(globalOpts, DEAL_COLUMNS)) + '\n',
      )
    })

  deal
    .command('show')
    .description('Show deal details')
    .argument('<id>', 'Deal ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createDealRepo(db)
      const cfRepo = createCustomFieldRepo(db)
      const id = Number(idStr)

      const deal = repo.findById(id)
      if (!deal) throw new NotFoundError('deal', id)

      const customFields = cfRepo.get('deal', id)

      const format = resolveFormat(globalOpts.format)
      if (format === 'json') {
        const data = { ...deal, custom_fields: customFields }
        process.stdout.write(JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0) + '\n')
      } else {
        process.stdout.write(
          formatOutput([deal], format, buildFormatOpts(globalOpts, DEAL_COLUMNS)) + '\n',
        )
        if (customFields.length > 0 && !globalOpts.quiet && format === 'table') {
          process.stdout.write('\nCustom Fields:\n')
          const cfData = customFields.map((cf) => ({
            field: cf.field_name,
            value: cf.field_value,
          }))
          process.stdout.write(formatOutput(cfData, 'table', buildFormatOpts(globalOpts)) + '\n')
        }
      }
    })

  deal
    .command('edit')
    .description('Edit a deal')
    .argument('<id>', 'Deal ID')
    .option('--title <title>', 'Deal title')
    .option('--value <amount>', 'Deal value')
    .option('--currency <code>', 'Currency code')
    .option('--stage <stage>', `Stage: ${DEAL_STAGES.join(', ')}`)
    .option('--person <id>', 'Person ID')
    .option('--org <id>', 'Organization ID')
    .option('--note <text>', 'Notes')
    .option('--closed-at <date>', 'Closed date')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createDealRepo(db)
      const id = Number(idStr)

      const existing = repo.findById(id)
      if (!existing) throw new NotFoundError('deal', id)

      const update: Record<string, unknown> = {}
      if (opts['title'] !== undefined) update['title'] = opts['title']
      if (opts['value'] !== undefined) update['value'] = Number(opts['value'])
      if (opts['currency'] !== undefined) update['currency'] = opts['currency']
      if (opts['stage'] !== undefined) update['stage'] = opts['stage']
      if (opts['person'] !== undefined) update['person_id'] = Number(opts['person'])
      if (opts['org'] !== undefined) update['org_id'] = Number(opts['org'])
      if (opts['note'] !== undefined) update['notes'] = opts['note']
      if (opts['closedAt'] !== undefined) update['closed_at'] = opts['closedAt']

      const deal = repo.update(id, update)
      if (!deal) throw new NotFoundError('deal', id)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([deal], format, buildFormatOpts(globalOpts, DEAL_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Updated deal #${String(id)}\n`)
      }
    })

  deal
    .command('delete')
    .description('Delete a deal (soft-delete)')
    .argument('<id>', 'Deal ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createDealRepo(db)
      const id = Number(idStr)

      const archived = repo.archive(id)
      if (!archived) throw new NotFoundError('deal', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Archived deal #${String(id)}\n`)
      }
    })

  deal
    .command('pipeline')
    .description('Show deal pipeline summary')
    .action((_opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createDealRepo(db)

      const summary = repo.pipeline()

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput(
          summary.map((s) => ({ ...s })),
          format,
          buildFormatOpts(globalOpts, PIPELINE_COLUMNS),
        ) + '\n',
      )
    })
}
