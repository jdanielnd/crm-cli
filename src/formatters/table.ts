import chalk from 'chalk'
import Table from 'cli-table3'

function toDisplayValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value === 'boolean') return value ? 'yes' : 'no'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value as string | number)
}

export interface ColumnDef {
  key: string
  header?: string
  align?: 'left' | 'right' | 'center'
  maxWidth?: number
}

function truncate(str: string, maxWidth: number): string {
  if (str.length <= maxWidth) return str
  return str.slice(0, maxWidth - 1) + '…'
}

export function formatTable(
  data: Record<string, unknown>[],
  columns?: (string | ColumnDef)[],
  options?: { noColor?: boolean },
): string {
  if (data.length === 0) return 'No results.'

  const cols: ColumnDef[] = (columns ?? Object.keys(data[0] ?? {}).map((k) => ({ key: k }))).map(
    (c) => (typeof c === 'string' ? { key: c } : c),
  )

  const table = new Table({
    head: cols.map((c) => {
      const label = c.header ?? c.key
      return options?.noColor ? label : chalk.bold(label)
    }),
    style: {
      head: [],
      border: [],
    },
  })

  for (const row of data) {
    table.push(
      cols.map((c) => {
        let val = toDisplayValue(row[c.key])
        if (c.maxWidth) val = truncate(val, c.maxWidth)
        return { content: val, hAlign: c.align ?? 'left' }
      }),
    )
  }

  return table.toString()
}
