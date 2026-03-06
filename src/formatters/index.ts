import { formatCsv, formatTsv } from './csv.js'
import { formatJson } from './json.js'
import { type ColumnDef, formatTable } from './table.js'

export type OutputFormat = 'table' | 'json' | 'csv' | 'tsv'

export type { ColumnDef } from './table.js'

export function resolveFormat(explicit?: string): OutputFormat {
  if (explicit) return explicit as OutputFormat
  return process.stdout.isTTY ? 'table' : 'json'
}

export interface FormatOptions {
  quiet?: boolean
  columns?: (string | ColumnDef)[]
  noColor?: boolean
}

export function formatOutput(
  data: Record<string, unknown>[],
  format: OutputFormat,
  options?: FormatOptions,
): string {
  if (options?.quiet) {
    return data.map((row) => row['id']).join('\n')
  }

  const colKeys = options?.columns?.map((c) => (typeof c === 'string' ? c : c.key))

  switch (format) {
    case 'json':
      return formatJson(data)
    case 'csv':
      return formatCsv(data, colKeys)
    case 'tsv':
      return formatTsv(data, colKeys)
    case 'table':
      return formatTable(data, options?.columns, options?.noColor ? { noColor: true } : undefined)
  }
}
