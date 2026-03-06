import { describe, expect, it } from 'vitest'

import {
  DealInsert,
  InteractionInsert,
  OrganizationInsert,
  PersonInsert,
  PersonUpdate,
  RelationshipInsert,
  TaskInsert,
} from '../../src/models/types.js'

describe('PersonInsert schema', () => {
  it('should accept valid input', () => {
    const result = PersonInsert.safeParse({
      first_name: 'Jane',
      last_name: 'Smith',
      email: 'jane@example.com',
    })
    expect(result.success).toBe(true)
  })

  it('should require first_name', () => {
    const result = PersonInsert.safeParse({})
    expect(result.success).toBe(false)
  })

  it('should reject empty first_name', () => {
    const result = PersonInsert.safeParse({ first_name: '' })
    expect(result.success).toBe(false)
  })

  it('should accept minimal input (first_name only)', () => {
    const result = PersonInsert.safeParse({ first_name: 'Jane' })
    expect(result.success).toBe(true)
  })
})

describe('PersonUpdate schema', () => {
  it('should accept partial updates', () => {
    const result = PersonUpdate.safeParse({ email: 'new@example.com' })
    expect(result.success).toBe(true)
  })

  it('should accept empty object', () => {
    const result = PersonUpdate.safeParse({})
    expect(result.success).toBe(true)
  })
})

describe('OrganizationInsert schema', () => {
  it('should accept valid input', () => {
    const result = OrganizationInsert.safeParse({
      name: 'Acme Corp',
      domain: 'acme.com',
      industry: 'SaaS',
    })
    expect(result.success).toBe(true)
  })

  it('should require name', () => {
    const result = OrganizationInsert.safeParse({})
    expect(result.success).toBe(false)
  })
})

describe('InteractionInsert schema', () => {
  it('should accept valid input', () => {
    const result = InteractionInsert.safeParse({
      type: 'call',
      subject: 'Discussed roadmap',
      direction: 'outbound',
    })
    expect(result.success).toBe(true)
  })

  it('should reject invalid type', () => {
    const result = InteractionInsert.safeParse({ type: 'invalid' })
    expect(result.success).toBe(false)
  })

  it('should reject invalid direction', () => {
    const result = InteractionInsert.safeParse({ type: 'call', direction: 'sideways' })
    expect(result.success).toBe(false)
  })
})

describe('DealInsert schema', () => {
  it('should accept valid input', () => {
    const result = DealInsert.safeParse({
      title: 'Website Redesign',
      value: 15000,
      stage: 'proposal',
    })
    expect(result.success).toBe(true)
  })

  it('should reject invalid stage', () => {
    const result = DealInsert.safeParse({ title: 'Deal', stage: 'invalid' })
    expect(result.success).toBe(false)
  })
})

describe('TaskInsert schema', () => {
  it('should accept valid input', () => {
    const result = TaskInsert.safeParse({
      title: 'Follow up',
      priority: 'high',
      due_at: '2026-03-15',
    })
    expect(result.success).toBe(true)
  })

  it('should reject invalid priority', () => {
    const result = TaskInsert.safeParse({ title: 'Task', priority: 'critical' })
    expect(result.success).toBe(false)
  })
})

describe('RelationshipInsert schema', () => {
  it('should accept valid input', () => {
    const result = RelationshipInsert.safeParse({
      person_id: 1,
      related_person_id: 2,
      type: 'colleague',
    })
    expect(result.success).toBe(true)
  })

  it('should reject invalid relationship type', () => {
    const result = RelationshipInsert.safeParse({
      person_id: 1,
      related_person_id: 2,
      type: 'enemy',
    })
    expect(result.success).toBe(false)
  })
})
