import { describe, expect, it } from 'vitest'

import { formatCsv, formatTsv } from '../../src/formatters/csv.js'
import { formatJson } from '../../src/formatters/json.js'
import { formatTable } from '../../src/formatters/table.js'
import { formatOutput } from '../../src/formatters/index.js'

const sampleData = [
  { id: 1, name: 'Jane Smith', email: 'jane@example.com' },
  { id: 2, name: 'Bob Jones', email: 'bob@example.com' },
]

describe('formatJson', () => {
  it('should produce valid JSON', () => {
    const result = formatJson(sampleData)
    const parsed: unknown = JSON.parse(result)
    expect(parsed).toEqual(sampleData)
  })

  it('should handle empty array', () => {
    expect(formatJson([])).toBe('[]')
  })

  it('should handle null values', () => {
    const data = [{ id: 1, name: null }]
    const parsed: unknown = JSON.parse(formatJson(data))
    expect(parsed).toEqual([{ id: 1, name: null }])
  })
})

describe('formatCsv', () => {
  it('should produce CSV with headers', () => {
    const result = formatCsv(sampleData)
    const lines = result.split('\n')
    expect(lines[0]).toBe('id,name,email')
    expect(lines[1]).toBe('1,Jane Smith,jane@example.com')
    expect(lines[2]).toBe('2,Bob Jones,bob@example.com')
  })

  it('should escape fields with commas', () => {
    const data = [{ id: 1, name: 'Smith, Jane' }]
    const result = formatCsv(data)
    expect(result).toContain('"Smith, Jane"')
  })

  it('should escape fields with quotes', () => {
    const data = [{ id: 1, name: 'Jane "JJ" Smith' }]
    const result = formatCsv(data)
    expect(result).toContain('"Jane ""JJ"" Smith"')
  })

  it('should handle null values as empty strings', () => {
    const data = [{ id: 1, name: null }]
    const result = formatCsv(data as Record<string, unknown>[])
    expect(result).toBe('id,name\n1,')
  })

  it('should return empty string for empty data', () => {
    expect(formatCsv([])).toBe('')
  })

  it('should respect column selection', () => {
    const result = formatCsv(sampleData, ['name', 'email'])
    const lines = result.split('\n')
    expect(lines[0]).toBe('name,email')
    expect(lines[1]).toBe('Jane Smith,jane@example.com')
  })
})

describe('formatTsv', () => {
  it('should use tab delimiter', () => {
    const result = formatTsv(sampleData)
    const lines = result.split('\n')
    expect(lines[0]).toBe('id\tname\temail')
    expect(lines[1]).toBe('1\tJane Smith\tjane@example.com')
  })
})

describe('formatTable', () => {
  it('should produce a table string', () => {
    const result = formatTable(sampleData, undefined, { noColor: true })
    expect(result).toContain('Jane Smith')
    expect(result).toContain('Bob Jones')
    expect(result).toContain('jane@example.com')
  })

  it('should return "No results." for empty data', () => {
    expect(formatTable([])).toBe('No results.')
  })

  it('should respect column selection', () => {
    const result = formatTable(sampleData, ['name'], { noColor: true })
    expect(result).toContain('Jane Smith')
    expect(result).not.toContain('jane@example.com')
  })

  it('should use custom headers', () => {
    const result = formatTable(sampleData, [{ key: 'name', header: 'Full Name' }], {
      noColor: true,
    })
    expect(result).toContain('Full Name')
  })
})

describe('formatOutput', () => {
  it('should output IDs in quiet mode', () => {
    const result = formatOutput(sampleData, 'json', { quiet: true })
    expect(result).toBe('1\n2')
  })

  it('should dispatch to json formatter', () => {
    const result = formatOutput(sampleData, 'json')
    const parsed: unknown = JSON.parse(result)
    expect(parsed).toEqual(sampleData)
  })

  it('should dispatch to csv formatter', () => {
    const result = formatOutput(sampleData, 'csv')
    expect(result).toContain('id,name,email')
  })

  it('should dispatch to tsv formatter', () => {
    const result = formatOutput(sampleData, 'tsv')
    expect(result).toContain('id\tname\temail')
  })

  it('should dispatch to table formatter', () => {
    const result = formatOutput(sampleData, 'table', { noColor: true })
    expect(result).toContain('Jane Smith')
  })
})
