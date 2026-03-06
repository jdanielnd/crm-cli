import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createCustomFieldRepo } from '../../db/repositories/custom-field.repo.js'
import { createPersonRepo } from '../../db/repositories/person.repo.js'
import {
  type ColumnDef,
  type FormatOptions,
  formatOutput,
  resolveFormat,
} from '../../formatters/index.js'
import { NotFoundError, ValidationError } from '../../models/errors.js'
import type { PersonFilters } from '../../db/repositories/person.repo.js'

const PERSON_COLUMNS: ColumnDef[] = [
  { key: 'id', header: 'ID' },
  { key: 'first_name', header: 'First' },
  { key: 'last_name', header: 'Last' },
  { key: 'email', header: 'Email' },
  { key: 'phone', header: 'Phone' },
  { key: 'title', header: 'Title' },
  { key: 'company', header: 'Company' },
]

function buildFormatOpts(globalOpts: GlobalOpts, columns?: ColumnDef[]): FormatOptions {
  const opts: FormatOptions = {}
  if (globalOpts.quiet !== undefined) opts.quiet = globalOpts.quiet
  if (globalOpts.color === false) opts.noColor = true
  if (columns) opts.columns = columns
  return opts
}

function parseName(name: string): { first_name: string; last_name: string | null } {
  const parts = name.trim().split(/\s+/)
  const first = parts[0]
  if (!first) throw new ValidationError('Name cannot be empty')
  const rest = parts.slice(1).join(' ')
  return { first_name: first, last_name: rest || null }
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

export function registerPersonCommands(program: Command): void {
  const person = program.command('person').description('Manage contacts')

  person
    .command('add')
    .description('Add a new person')
    .argument('<name>', 'Full name (e.g. "Jane Smith")')
    .option('--email <email>', 'Email address')
    .option('--phone <phone>', 'Phone number')
    .option('--title <title>', 'Job title')
    .option('--company <company>', 'Company name')
    .option('--location <location>', 'Location')
    .option('--linkedin <url>', 'LinkedIn URL')
    .option('--twitter <handle>', 'Twitter handle')
    .option('--website <url>', 'Website URL')
    .option('--note <text>', 'Notes')
    .option('--org <id>', 'Organization ID')
    .option('--set <key=value...>', 'Set custom fields')
    .action((name: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createPersonRepo(db)
      const cfRepo = createCustomFieldRepo(db)
      const { first_name, last_name } = parseName(name)

      const person = repo.create({
        first_name,
        last_name,
        email: (opts['email'] as string | undefined) ?? null,
        phone: (opts['phone'] as string | undefined) ?? null,
        title: (opts['title'] as string | undefined) ?? null,
        company: (opts['company'] as string | undefined) ?? null,
        location: (opts['location'] as string | undefined) ?? null,
        linkedin: (opts['linkedin'] as string | undefined) ?? null,
        twitter: (opts['twitter'] as string | undefined) ?? null,
        website: (opts['website'] as string | undefined) ?? null,
        notes: (opts['note'] as string | undefined) ?? null,
        org_id: opts['org'] ? Number(opts['org']) : null,
      })

      const customFields = parseSetFlags(opts['set'] as string[] | undefined)
      for (const [key, value] of Object.entries(customFields)) {
        cfRepo.set('person', person.id, key, value)
      }

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([person], format, buildFormatOpts(globalOpts, PERSON_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Created person #${String(person.id)}\n`)
      }
    })

  person
    .command('list')
    .description('List people')
    .option('--tag <tag>', 'Filter by tag')
    .option('--org <id>', 'Filter by organization ID')
    .option('--search <query>', 'Search by name, email, etc.')
    .option('--sort <field>', 'Sort by: name, created, updated')
    .option('--limit <n>', 'Max results', '100')
    .action((opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createPersonRepo(db)

      const filters: PersonFilters = { limit: Number(opts['limit']) }
      if (opts['tag']) filters.tag = opts['tag'] as string
      if (opts['org']) filters.orgId = Number(opts['org'])
      if (opts['search']) filters.search = opts['search'] as string
      if (opts['sort']) filters.sort = opts['sort'] as 'name' | 'created' | 'updated'

      const people = repo.findAll(filters)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput(people, format, buildFormatOpts(globalOpts, PERSON_COLUMNS)) + '\n',
      )
    })

  person
    .command('show')
    .description('Show person details')
    .argument('<id>', 'Person ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createPersonRepo(db)
      const cfRepo = createCustomFieldRepo(db)
      const id = Number(idStr)

      const person = repo.findById(id)
      if (!person) throw new NotFoundError('person', id)

      const customFields = cfRepo.get('person', id)

      const format = resolveFormat(globalOpts.format)
      if (format === 'json') {
        const data = { ...person, custom_fields: customFields }
        process.stdout.write(JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0) + '\n')
      } else {
        process.stdout.write(formatOutput([person], format, buildFormatOpts(globalOpts)) + '\n')
        if (customFields.length > 0 && !globalOpts.quiet && format === 'table') {
          process.stdout.write('\nCustom Fields:\n')
          const cfData = customFields.map((cf) => ({ field: cf.field_name, value: cf.field_value }))
          process.stdout.write(formatOutput(cfData, 'table', buildFormatOpts(globalOpts)) + '\n')
        }
      }
    })

  person
    .command('edit')
    .description('Edit a person')
    .argument('<id>', 'Person ID')
    .option('--first-name <name>', 'First name')
    .option('--last-name <name>', 'Last name')
    .option('--email <email>', 'Email address')
    .option('--phone <phone>', 'Phone number')
    .option('--title <title>', 'Job title')
    .option('--company <company>', 'Company name')
    .option('--location <location>', 'Location')
    .option('--linkedin <url>', 'LinkedIn URL')
    .option('--twitter <handle>', 'Twitter handle')
    .option('--website <url>', 'Website URL')
    .option('--note <text>', 'Notes')
    .option('--summary <text>', 'AI summary')
    .option('--org <id>', 'Organization ID')
    .option('--set <key=value...>', 'Set custom fields')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createPersonRepo(db)
      const cfRepo = createCustomFieldRepo(db)
      const id = Number(idStr)

      const existing = repo.findById(id)
      if (!existing) throw new NotFoundError('person', id)

      const update: Record<string, unknown> = {}
      if (opts['firstName'] !== undefined) update['first_name'] = opts['firstName']
      if (opts['lastName'] !== undefined) update['last_name'] = opts['lastName']
      if (opts['email'] !== undefined) update['email'] = opts['email']
      if (opts['phone'] !== undefined) update['phone'] = opts['phone']
      if (opts['title'] !== undefined) update['title'] = opts['title']
      if (opts['company'] !== undefined) update['company'] = opts['company']
      if (opts['location'] !== undefined) update['location'] = opts['location']
      if (opts['linkedin'] !== undefined) update['linkedin'] = opts['linkedin']
      if (opts['twitter'] !== undefined) update['twitter'] = opts['twitter']
      if (opts['website'] !== undefined) update['website'] = opts['website']
      if (opts['note'] !== undefined) update['notes'] = opts['note']
      if (opts['summary'] !== undefined) update['summary'] = opts['summary']
      if (opts['org'] !== undefined) update['org_id'] = Number(opts['org'])

      const person = repo.update(id, update)
      if (!person) throw new NotFoundError('person', id)

      const customFields = parseSetFlags(opts['set'] as string[] | undefined)
      for (const [key, value] of Object.entries(customFields)) {
        cfRepo.set('person', id, key, value)
      }

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([person], format, buildFormatOpts(globalOpts, PERSON_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Updated person #${String(id)}\n`)
      }
    })

  person
    .command('delete')
    .description('Delete a person (soft-delete)')
    .argument('<id>', 'Person ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createPersonRepo(db)
      const id = Number(idStr)

      const archived = repo.archive(id)
      if (!archived) throw new NotFoundError('person', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Archived person #${String(id)}\n`)
      }
    })
}
