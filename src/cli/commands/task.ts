import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { createTaskRepo, type TaskFilters } from '../../db/repositories/task.repo.js'
import {
  type ColumnDef,
  type FormatOptions,
  formatOutput,
  resolveFormat,
} from '../../formatters/index.js'
import { NotFoundError } from '../../models/errors.js'
import { PRIORITIES, type Priority } from '../../models/types.js'

const TASK_COLUMNS: ColumnDef[] = [
  { key: 'id', header: 'ID' },
  { key: 'title', header: 'Title' },
  { key: 'due_at', header: 'Due' },
  { key: 'priority', header: 'Priority' },
  { key: 'completed', header: 'Done' },
  { key: 'person_id', header: 'Person' },
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

export function registerTaskCommands(program: Command): void {
  const task = program.command('task').description('Manage tasks & follow-ups')

  task
    .command('add')
    .description('Add a new task')
    .argument('<title>', 'Task title')
    .option('--person <id>', 'Person ID')
    .option('--deal <id>', 'Deal ID')
    .option('--due <date>', 'Due date')
    .option('--priority <level>', `Priority: ${PRIORITIES.join(', ')}`, 'normal')
    .option('--description <text>', 'Description')
    .action((title: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTaskRepo(db)

      const task = repo.create({
        title,
        description: (opts['description'] as string | undefined) ?? null,
        due_at: (opts['due'] as string | undefined) ?? null,
        priority: opts['priority'] as Priority | undefined,
        person_id: opts['person'] ? Number(opts['person']) : null,
        deal_id: opts['deal'] ? Number(opts['deal']) : null,
      })

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([task], format, buildFormatOpts(globalOpts, TASK_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Created task #${String(task.id)}\n`)
      }
    })

  task
    .command('list')
    .description('List tasks')
    .option('--person <id>', 'Filter by person ID')
    .option('--deal <id>', 'Filter by deal ID')
    .option('--priority <level>', 'Filter by priority')
    .option('--overdue', 'Show only overdue tasks')
    .option('--done', 'Show completed tasks')
    .option('--pending', 'Show only pending tasks')
    .option('--sort <field>', 'Sort by: due, priority, created')
    .option('--limit <n>', 'Max results', '100')
    .action((opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTaskRepo(db)

      const filters: TaskFilters = { limit: Number(opts['limit']) }
      if (opts['person']) filters.personId = Number(opts['person'])
      if (opts['deal']) filters.dealId = Number(opts['deal'])
      if (opts['priority']) filters.priority = opts['priority'] as Priority
      if (opts['overdue']) filters.overdue = true
      if (opts['done']) filters.completed = true
      if (opts['pending']) filters.completed = false
      if (opts['sort']) filters.sort = opts['sort'] as 'due' | 'priority' | 'created'

      const tasks = repo.findAll(filters)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput(tasks, format, buildFormatOpts(globalOpts, TASK_COLUMNS)) + '\n',
      )
    })

  task
    .command('show')
    .description('Show task details')
    .argument('<id>', 'Task ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTaskRepo(db)
      const id = Number(idStr)

      const task = repo.findById(id)
      if (!task) throw new NotFoundError('task', id)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(formatOutput([task], format, buildFormatOpts(globalOpts)) + '\n')
    })

  task
    .command('edit')
    .description('Edit a task')
    .argument('<id>', 'Task ID')
    .option('--title <title>', 'Task title')
    .option('--description <text>', 'Description')
    .option('--due <date>', 'Due date')
    .option('--priority <level>', `Priority: ${PRIORITIES.join(', ')}`)
    .option('--person <id>', 'Person ID')
    .option('--deal <id>', 'Deal ID')
    .action((idStr: string, opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTaskRepo(db)
      const id = Number(idStr)

      const existing = repo.findById(id)
      if (!existing) throw new NotFoundError('task', id)

      const update: Record<string, unknown> = {}
      if (opts['title'] !== undefined) update['title'] = opts['title']
      if (opts['description'] !== undefined) update['description'] = opts['description']
      if (opts['due'] !== undefined) update['due_at'] = opts['due']
      if (opts['priority'] !== undefined) update['priority'] = opts['priority']
      if (opts['person'] !== undefined) update['person_id'] = Number(opts['person'])
      if (opts['deal'] !== undefined) update['deal_id'] = Number(opts['deal'])

      const task = repo.update(id, update)
      if (!task) throw new NotFoundError('task', id)

      const format = resolveFormat(globalOpts.format)
      process.stdout.write(
        formatOutput([task], format, buildFormatOpts(globalOpts, TASK_COLUMNS)) + '\n',
      )
      if (!globalOpts.quiet) {
        process.stderr.write(`Updated task #${String(id)}\n`)
      }
    })

  task
    .command('done')
    .description('Mark a task as completed')
    .argument('<id>', 'Task ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTaskRepo(db)
      const id = Number(idStr)

      const task = repo.complete(id)
      if (!task) throw new NotFoundError('task', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Completed task #${String(id)}\n`)
      }
    })

  task
    .command('delete')
    .description('Delete a task (soft-delete)')
    .argument('<id>', 'Task ID')
    .action((idStr: string, _opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)
      const repo = createTaskRepo(db)
      const id = Number(idStr)

      const archived = repo.archive(id)
      if (!archived) throw new NotFoundError('task', id)

      if (!globalOpts.quiet) {
        process.stderr.write(`Archived task #${String(id)}\n`)
      }
    })
}
