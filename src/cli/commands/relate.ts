import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createRelationshipRepo } from '../../db/repositories/relationship.repo.js'
import { type FormatOptions, formatOutput, resolveFormat } from '../../formatters/index.js'
import { NotFoundError, ValidationError } from '../../models/errors.js'
import { RELATIONSHIP_TYPES, type RelationshipType } from '../../models/types.js'

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

function validateRelType(type: string): RelationshipType {
  if (!(RELATIONSHIP_TYPES as readonly string[]).includes(type)) {
    throw new ValidationError(
      `Invalid relationship type: "${type}". Use: ${RELATIONSHIP_TYPES.join(', ')}`,
    )
  }
  return type as RelationshipType
}

export function registerRelateCommands(program: Command): void {
  const person = program.commands.find((c) => c.name() === 'person')
  if (!person) return

  person
    .command('relate')
    .description('Create a relationship between two people')
    .argument('<id>', 'Person ID')
    .argument('<related-id>', 'Related person ID')
    .option('--type <type>', `Relationship type: ${RELATIONSHIP_TYPES.join(', ')}`, 'colleague')
    .option('--note <text>', 'Notes about the relationship')
    .action((idStr: string, relatedIdStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createRelationshipRepo(db)

      const personId = Number(idStr)
      const relatedPersonId = Number(relatedIdStr)
      const type = validateRelType(opts['type'] as string)

      const rel = repo.create({
        person_id: personId,
        related_person_id: relatedPersonId,
        type,
        notes: (opts['note'] as string | undefined) ?? null,
      })

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(formatOutput([rel], format, buildFormatOpts(globalOpts)) + '\n')
      if (!globalOpts.quiet) {
        process.stderr.write(
          `Created ${type} relationship between #${String(personId)} and #${String(relatedPersonId)}\n`,
        )
      }
    })

  person
    .command('relationships')
    .description('List relationships for a person')
    .argument('<id>', 'Person ID')
    .option('--type <type>', 'Filter by relationship type')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createRelationshipRepo(db)

      const personId = Number(idStr)

      let rels
      if (opts['type']) {
        const type = validateRelType(opts['type'] as string)
        rels = repo.findByType(personId, type)
      } else {
        rels = repo.findForPerson(personId)
      }

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(formatOutput(rels, format, buildFormatOpts(globalOpts)) + '\n')
    })

  person
    .command('unrelate')
    .description('Remove a relationship')
    .argument('<relationship-id>', 'Relationship ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createRelationshipRepo(db)

      const id = Number(idStr)
      const deleted = repo.delete(id)
      if (!deleted) throw new NotFoundError('relationship', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Deleted relationship #${String(id)}\n`)
      }
    })
}
