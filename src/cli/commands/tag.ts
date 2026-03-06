import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createTagRepo } from '../../db/repositories/tag.repo.js'
import { type FormatOptions, formatOutput, resolveFormat } from '../../formatters/index.js'
import { ValidationError } from '../../models/errors.js'
import { TAGGABLE_ENTITIES, type TaggableEntity } from '../../models/types.js'

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

function validateEntityType(type: string): TaggableEntity {
  if (!(TAGGABLE_ENTITIES as readonly string[]).includes(type)) {
    throw new ValidationError(
      `Invalid entity type: "${type}". Use: ${TAGGABLE_ENTITIES.join(', ')}`,
    )
  }
  return type as TaggableEntity
}

export function registerTagCommands(program: Command): void {
  const tag = program.command('tag').description('Manage tags')

  tag
    .command('list')
    .description('List all tags')
    .action((_opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTagRepo(db)

      const tags = repo.findAll()

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(formatOutput(tags, format, buildFormatOpts(globalOpts)) + '\n')
    })

  tag
    .command('apply')
    .description('Apply a tag to an entity')
    .argument('<entity-type>', `Entity type: ${TAGGABLE_ENTITIES.join(', ')}`)
    .argument('<entity-id>', 'Entity ID')
    .argument('<tag-name>', 'Tag name')
    .action(
      (
        entityType: string,
        entityIdStr: string,
        tagName: string,
        _opts: Record<string, unknown>,
        cmd: Command,
      ) => {
        const globalOpts = getOpts(cmd)
        const db = getDb(globalOpts.db)
        const repo = createTagRepo(db)

        const type = validateEntityType(entityType)
        const entityId = Number(entityIdStr)

        repo.apply(type, entityId, tagName)

        if (!globalOpts.quiet) {
          process.stderr.write(`Tagged ${type} #${String(entityId)} as "${tagName}"\n`)
        }
      },
    )

  tag
    .command('remove')
    .description('Remove a tag from an entity')
    .argument('<entity-type>', `Entity type: ${TAGGABLE_ENTITIES.join(', ')}`)
    .argument('<entity-id>', 'Entity ID')
    .argument('<tag-name>', 'Tag name')
    .action(
      (
        entityType: string,
        entityIdStr: string,
        tagName: string,
        _opts: Record<string, unknown>,
        cmd: Command,
      ) => {
        const globalOpts = getOpts(cmd)
        const db = getDb(globalOpts.db)
        const repo = createTagRepo(db)

        const type = validateEntityType(entityType)
        const entityId = Number(entityIdStr)

        const removed = repo.remove(type, entityId, tagName)

        if (!globalOpts.quiet) {
          if (removed) {
            process.stderr.write(`Removed tag "${tagName}" from ${type} #${String(entityId)}\n`)
          } else {
            process.stderr.write(`Tag "${tagName}" was not on ${type} #${String(entityId)}\n`)
          }
        }
      },
    )

  tag
    .command('show')
    .description('Show tags for an entity')
    .argument('<entity-type>', `Entity type: ${TAGGABLE_ENTITIES.join(', ')}`)
    .argument('<entity-id>', 'Entity ID')
    .action(
      (entityType: string, entityIdStr: string, _opts: Record<string, unknown>, cmd: Command) => {
        const globalOpts = getOpts(cmd)
        const db = getDb(globalOpts.db)
        const repo = createTagRepo(db)

        const type = validateEntityType(entityType)
        const entityId = Number(entityIdStr)

        const tags = repo.getForEntity(type, entityId)

        const format = resolveFormat(globalOpts.format)
        process.stdout.write(formatOutput(tags, format, buildFormatOpts(globalOpts)) + '\n')
      },
    )

  tag
    .command('delete')
    .description('Delete a tag entirely')
    .argument('<tag-name>', 'Tag name')
    .action((tagName: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTagRepo(db)

      const deleted = repo.delete(tagName)

      if (!globalOpts.quiet) {
        if (deleted) {
          process.stderr.write(`Deleted tag "${tagName}"\n`)
        } else {
          process.stderr.write(`Tag "${tagName}" not found\n`)
        }
      }
    })
}
