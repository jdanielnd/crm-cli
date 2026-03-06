import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createCustomFieldRepo } from '../../db/repositories/custom-field.repo.js'
import { createOrgRepo, type OrgFilters } from '../../db/repositories/org.repo.js'
import { createPersonRepo } from '../../db/repositories/person.repo.js'
import {
  type ColumnDef,
  type FormatOptions,
  formatOutput,
  resolveFormat,
} from '../../formatters/index.js'
import { NotFoundError, ValidationError } from '../../models/errors.js'

const ORG_COLUMNS: ColumnDef[] = [
  { key: 'id', header: 'ID' },
  { key: 'name', header: 'Name' },
  { key: 'domain', header: 'Domain' },
  { key: 'industry', header: 'Industry' },
  { key: 'location', header: 'Location' },
]

interface GlobalOpts {
  format?: string
  quiet?: boolean
  db?: string
  color?: boolean
}

function getOpts(cmd: Command): GlobalOpts {
  const root = cmd.parent ?? cmd
  return root.opts()
}

function buildFormatOpts(globalOpts: GlobalOpts, columns?: ColumnDef[]): FormatOptions {
  const opts: FormatOptions = {}
  if (globalOpts.quiet !== undefined) opts.quiet = globalOpts.quiet
  if (globalOpts.color === false) opts.noColor = true
  if (columns) opts.columns = columns
  return opts
}

function parseSetFlags(setFlags: string[] | undefined): Record<string, string> {
  const fields: Record<string, string> = {}
  if (!setFlags) return fields
  for (const flag of setFlags) {
    const eqIndex = flag.indexOf('=')
    if (eqIndex === -1) throw new ValidationError(`Invalid --set format: "${flag}". Use key=value`)
    const key = flag.slice(0, eqIndex)
    const value = flag.slice(eqIndex + 1)
    if (!key) throw new ValidationError(`Invalid --set format: "${flag}". Key cannot be empty`)
    fields[key] = value
  }
  return fields
}

export function registerOrgCommands(program: Command): void {
  const org = program.command('org').description('Manage organizations')

  org
    .command('add')
    .description('Add a new organization')
    .argument('<name>', 'Organization name')
    .option('--domain <domain>', 'Domain (e.g. acme.com)')
    .option('--industry <industry>', 'Industry')
    .option('--location <location>', 'Location')
    .option('--note <text>', 'Notes')
    .option('--set <key=value...>', 'Set custom fields')
    .action((name: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createOrgRepo(db)
      const cfRepo = createCustomFieldRepo(db)

      if (!name.trim()) throw new ValidationError('Organization name cannot be empty')

      const organization = repo.create({
        name: name.trim(),
        domain: (opts['domain'] as string | undefined) ?? null,
        industry: (opts['industry'] as string | undefined) ?? null,
        location: (opts['location'] as string | undefined) ?? null,
        notes: (opts['note'] as string | undefined) ?? null,
      })

      const customFields = parseSetFlags(opts['set'] as string[] | undefined)
      for (const [key, value] of Object.entries(customFields)) {
        cfRepo.set('organization', organization.id, key, value)
      }

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([organization], format, buildFormatOpts(globalOpts, ORG_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Created organization #${String(organization.id)}\n`)
      }
    })

  org
    .command('list')
    .description('List organizations')
    .option('--search <query>', 'Search by name, domain, etc.')
    .option('--sort <field>', 'Sort by: name, created, updated')
    .option('--limit <n>', 'Max results', '100')
    .action((opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createOrgRepo(db)

      const filters: OrgFilters = { limit: Number(opts['limit']) }
      if (opts['search']) filters.search = opts['search'] as string
      if (opts['sort']) filters.sort = opts['sort'] as 'name' | 'created' | 'updated'

      const orgs = repo.findAll(filters)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput(orgs, format, buildFormatOpts(globalOpts, ORG_COLUMNS)) + '\n',
      )
    })

  org
    .command('show')
    .description('Show organization details')
    .argument('<id>', 'Organization ID')
    .option('--with <related>', 'Include related: people')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createOrgRepo(db)
      const cfRepo = createCustomFieldRepo(db)
      const id = Number(idStr)

      const organization = repo.findById(id)
      if (!organization) throw new NotFoundError('organization', id)

      const customFields = cfRepo.get('organization', id)

      const format = resolveFormat(globalOpts.format)
      if (format === 'json') {
        const data: Record<string, unknown> = { ...organization, custom_fields: customFields }
        if (opts['with'] === 'people') {
          const personRepo = createPersonRepo(db)
          data['people'] = personRepo.findAll({ orgId: id })
        }
        process.stdout.write(JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0) + '\n')
      } else {
        process.stdout.write(
          formatOutput([organization], format, buildFormatOpts(globalOpts)) + '\n',
        )
        if (customFields.length > 0 && !globalOpts.quiet && format === 'table') {
          process.stdout.write('\nCustom Fields:\n')
          const cfData = customFields.map((cf) => ({
            field: cf.field_name,
            value: cf.field_value,
          }))
          process.stdout.write(formatOutput(cfData, 'table', buildFormatOpts(globalOpts)) + '\n')
        }
        if (opts['with'] === 'people' && format === 'table') {
          const personRepo = createPersonRepo(db)
          const people = personRepo.findAll({ orgId: id })
          if (people.length > 0) {
            process.stdout.write('\nPeople:\n')
            process.stdout.write(formatOutput(people, 'table', buildFormatOpts(globalOpts)) + '\n')
          }
        }
      }
    })

  org
    .command('edit')
    .description('Edit an organization')
    .argument('<id>', 'Organization ID')
    .option('--name <name>', 'Organization name')
    .option('--domain <domain>', 'Domain')
    .option('--industry <industry>', 'Industry')
    .option('--location <location>', 'Location')
    .option('--note <text>', 'Notes')
    .option('--summary <text>', 'AI summary')
    .option('--set <key=value...>', 'Set custom fields')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createOrgRepo(db)
      const cfRepo = createCustomFieldRepo(db)
      const id = Number(idStr)

      const existing = repo.findById(id)
      if (!existing) throw new NotFoundError('organization', id)

      const update: Record<string, unknown> = {}
      if (opts['name'] !== undefined) update['name'] = opts['name']
      if (opts['domain'] !== undefined) update['domain'] = opts['domain']
      if (opts['industry'] !== undefined) update['industry'] = opts['industry']
      if (opts['location'] !== undefined) update['location'] = opts['location']
      if (opts['note'] !== undefined) update['notes'] = opts['note']
      if (opts['summary'] !== undefined) update['summary'] = opts['summary']

      const organization = repo.update(id, update)
      if (!organization) throw new NotFoundError('organization', id)

      const customFields = parseSetFlags(opts['set'] as string[] | undefined)
      for (const [key, value] of Object.entries(customFields)) {
        cfRepo.set('organization', id, key, value)
      }

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([organization], format, buildFormatOpts(globalOpts, ORG_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Updated organization #${String(id)}\n`)
      }
    })

  org
    .command('delete')
    .description('Delete an organization (soft-delete)')
    .argument('<id>', 'Organization ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createOrgRepo(db)
      const id = Number(idStr)

      const archived = repo.archive(id)
      if (!archived) throw new NotFoundError('organization', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Archived organization #${String(id)}\n`)
      }
    })
}
