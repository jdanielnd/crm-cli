import { Command } from 'commander'

import { getDb } from '../../db/index.js'
import { resolveFormat } from '../../formatters/index.js'

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

type StatsRow = { count: number }
type DealStatsRow = { count: number; total_value: number }

export function registerStatusCommands(program: Command): void {
  program
    .command('status')
    .description('CRM summary dashboard')
    .action((_opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const db = getDb(globalOpts.db)

      const people = (
        db.prepare('SELECT COUNT(*) as count FROM people WHERE archived = 0').get() as StatsRow
      ).count
      const orgs = (
        db
          .prepare('SELECT COUNT(*) as count FROM organizations WHERE archived = 0')
          .get() as StatsRow
      ).count
      const openDeals = db
        .prepare(
          "SELECT COUNT(*) as count, COALESCE(SUM(value), 0) as total_value FROM deals WHERE archived = 0 AND stage NOT IN ('won', 'lost')",
        )
        .get() as DealStatsRow
      const overdueTasks = (
        db
          .prepare(
            "SELECT COUNT(*) as count FROM tasks WHERE archived = 0 AND completed = 0 AND due_at < datetime('now')",
          )
          .get() as StatsRow
      ).count
      const interactionsThisWeek = (
        db
          .prepare(
            "SELECT COUNT(*) as count FROM interactions WHERE archived = 0 AND occurred_at >= datetime('now', '-7 days')",
          )
          .get() as StatsRow
      ).count

      const format = resolveFormat(globalOpts.format)

      if (format === 'json') {
        const data = {
          people,
          organizations: orgs,
          open_deals: openDeals.count,
          open_deals_value: openDeals.total_value,
          overdue_tasks: overdueTasks,
          interactions_this_week: interactionsThisWeek,
        }
        process.stdout.write(JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0) + '\n')
      } else {
        const dealValue = openDeals.total_value
          ? ` ($${String(openDeals.total_value).replace(/\B(?=(\d{3})+(?!\d))/g, ',')})`
          : ''
        const lines = [
          `${String(people)} contacts | ${String(orgs)} organizations | ${String(openDeals.count)} open deals${dealValue}`,
          `${String(overdueTasks)} tasks overdue | ${String(interactionsThisWeek)} interactions this week`,
        ]
        process.stdout.write(lines.join('\n') + '\n')
      }
    })
}
