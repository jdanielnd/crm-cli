import { execSync } from 'node:child_process'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { describe, expect, it, beforeEach, afterEach } from 'vitest'

let tmpDir: string
let dbPath: string

function crm(args: string): { stdout: string; stderr: string; exitCode: number } {
  try {
    const result = execSync(`npx tsx src/cli/index.ts --db "${dbPath}" ${args}`, {
      encoding: 'utf-8',
      cwd: process.cwd(),
      env: { ...process.env, NODE_ENV: 'test' },
    })
    return { stdout: result, stderr: '', exitCode: 0 }
  } catch (error: unknown) {
    const e = error as { stdout?: string; stderr?: string; status?: number }
    return { stdout: e.stdout ?? '', stderr: e.stderr ?? '', exitCode: e.status ?? 1 }
  }
}

describe('CLI Integration', () => {
  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), 'crm-test-'))
    dbPath = join(tmpDir, 'crm.db')
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  it('should show version', () => {
    const { stdout } = crm('--version')
    expect(stdout.trim()).toMatch(/^\d+\.\d+\.\d+$/)
  })

  it('should show help', () => {
    const { stdout } = crm('--help')
    expect(stdout).toContain('Local-first personal CRM')
    expect(stdout).toContain('person')
    expect(stdout).toContain('org')
    expect(stdout).toContain('log')
    expect(stdout).toContain('deal')
    expect(stdout).toContain('task')
    expect(stdout).toContain('tag')
    expect(stdout).toContain('search')
    expect(stdout).toContain('context')
    expect(stdout).toContain('status')
    expect(stdout).toContain('mcp')
  })

  it('should add and list a person', () => {
    const add = crm('person add "Jane Smith" --email jane@example.com -f json')
    const person = JSON.parse(add.stdout) as Array<{ id: number; first_name: string }>
    expect(person[0]?.first_name).toBe('Jane')

    const list = crm('person list -f json')
    const people = JSON.parse(list.stdout) as Array<{ first_name: string }>
    expect(people).toHaveLength(1)
    expect(people[0]?.first_name).toBe('Jane')
  })

  it('should show a person', () => {
    crm('person add "Jane Smith" -f json')
    const show = crm('person show 1 -f json')
    const data = JSON.parse(show.stdout) as { first_name: string; custom_fields: unknown[] }
    expect(data.first_name).toBe('Jane')
    expect(data.custom_fields).toEqual([])
  })

  it('should edit a person', () => {
    crm('person add "Jane Smith" -f json')
    const edit = crm('person edit 1 --email new@example.com -f json')
    const person = JSON.parse(edit.stdout) as Array<{ email: string }>
    expect(person[0]?.email).toBe('new@example.com')
  })

  it('should delete a person', () => {
    crm('person add "Jane Smith" -f json')
    crm('person delete 1')
    const list = crm('person list -f json')
    const people = JSON.parse(list.stdout) as unknown[]
    expect(people).toHaveLength(0)
  })

  it('should add and list an org', () => {
    const add = crm('org add "Acme Corp" --domain acme.com -f json')
    const org = JSON.parse(add.stdout) as Array<{ name: string }>
    expect(org[0]?.name).toBe('Acme Corp')
  })

  it('should log an interaction', () => {
    crm('person add "Jane Smith" -f json')
    const log = crm('log call 1 --subject "Catch-up" -f json')
    const interaction = JSON.parse(log.stdout) as Array<{ type: string }>
    expect(interaction[0]?.type).toBe('call')
  })

  it('should manage tags', () => {
    crm('person add "Jane Smith" -f json')
    crm('tag apply person 1 vip')
    const tags = crm('tag show person 1 -f json')
    const data = JSON.parse(tags.stdout) as Array<{ name: string }>
    expect(data).toHaveLength(1)
    expect(data[0]?.name).toBe('vip')
  })

  it('should add and list deals', () => {
    crm('deal add "Website" --value 15000 --stage proposal -f json')
    const pipeline = crm('deal pipeline -f json')
    const data = JSON.parse(pipeline.stdout) as Array<{ stage: string; total_value: number }>
    expect(data[0]?.stage).toBe('proposal')
    expect(data[0]?.total_value).toBe(15000)
  })

  it('should add and complete tasks', () => {
    crm('task add "Follow up" --priority high -f json')
    crm('task done 1')
    const list = crm('task list --pending -f json')
    const tasks = JSON.parse(list.stdout) as unknown[]
    expect(tasks).toHaveLength(0)
  })

  it('should search across entities', () => {
    crm('person add "Jane Smith" -f json')
    crm('org add "Acme Corp" -f json')
    const search = crm('search Jane -f json')
    const results = JSON.parse(search.stdout) as Array<{ type: string }>
    expect(results.some((r) => r.type === 'person')).toBe(true)
  })

  it('should show status', () => {
    crm('person add "Jane Smith" -f json')
    crm('org add "Acme" -f json')
    const status = crm('status -f json')
    const data = JSON.parse(status.stdout) as { people: number; organizations: number }
    expect(data.people).toBe(1)
    expect(data.organizations).toBe(1)
  })

  it('should show context for a person', () => {
    crm('person add "Jane Smith" --email jane@example.com -f json')
    crm('log call 1 --subject "Test call" -f json')
    const ctx = crm('context 1 -f json')
    const data = JSON.parse(ctx.stdout) as {
      person: { first_name: string }
      recent_interactions: unknown[]
    }
    expect(data.person.first_name).toBe('Jane')
    expect(data.recent_interactions).toHaveLength(1)
  })

  it('should output quiet mode (IDs only)', () => {
    crm('person add "Jane Smith" -f json')
    crm('person add "Bob Jones" -f json')
    const { stdout } = crm('person list -q')
    const ids = stdout.trim().split('\n')
    expect(ids).toHaveLength(2)
  })

  it('should output table format', () => {
    crm('person add "Jane Smith" -f json')
    const { stdout } = crm('person list -f table')
    expect(stdout).toContain('Jane')
    expect(stdout).toContain('ID')
  })

  it('should output CSV format', () => {
    crm('person add "Jane Smith" --email jane@example.com -f json')
    const { stdout } = crm('person list -f csv')
    expect(stdout).toContain('jane@example.com')
  })
})
