function escapeField(value: string, delimiter: string): string {
  if (value.includes(delimiter) || value.includes('"') || value.includes('\n')) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}

function toStringValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value as string | number | boolean)
}

export function formatCsv(
  data: Record<string, unknown>[],
  columns?: string[],
  delimiter = ',',
): string {
  if (data.length === 0) return ''

  const cols = columns ?? Object.keys(data[0] ?? {})
  const header = cols.map((c) => escapeField(c, delimiter)).join(delimiter)

  const rows = data.map((row) =>
    cols.map((col) => escapeField(toStringValue(row[col]), delimiter)).join(delimiter),
  )

  return [header, ...rows].join('\n')
}

export function formatTsv(data: Record<string, unknown>[], columns?: string[]): string {
  return formatCsv(data, columns, '\t')
}
