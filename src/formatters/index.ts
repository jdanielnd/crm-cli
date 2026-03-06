export type OutputFormat = 'table' | 'json' | 'csv' | 'tsv'

export function resolveFormat(explicit?: string): OutputFormat {
  if (explicit) return explicit as OutputFormat
  return process.stdout.isTTY ? 'table' : 'json'
}

export function formatOutput(
  data: Record<string, unknown>[],
  format: OutputFormat,
  options?: { quiet?: boolean; columns?: string[] },
): string {
  if (options?.quiet) {
    return data.map((row) => row['id']).join('\n')
  }

  switch (format) {
    case 'json':
      return JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0)
    case 'csv':
    case 'tsv':
    case 'table':
    default:
      // Placeholder — will be implemented in Step 2
      return JSON.stringify(data, null, 2)
  }
}
