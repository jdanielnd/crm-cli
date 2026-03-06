import { describe, expect, it } from 'vitest'

import {
  CliError,
  ConflictError,
  DatabaseError,
  NotFoundError,
  ValidationError,
} from '../../src/models/errors.js'

describe('CliError', () => {
  it('should have default exit code 1', () => {
    const err = new CliError('something went wrong')
    expect(err.exitCode).toBe(1)
    expect(err.message).toBe('something went wrong')
    expect(err.name).toBe('CliError')
  })

  it('should accept custom exit code', () => {
    const err = new CliError('custom', 42)
    expect(err.exitCode).toBe(42)
  })
})

describe('ValidationError', () => {
  it('should have exit code 2', () => {
    const err = new ValidationError('invalid input')
    expect(err.exitCode).toBe(2)
    expect(err.name).toBe('ValidationError')
  })
})

describe('NotFoundError', () => {
  it('should have exit code 3 and format message', () => {
    const err = new NotFoundError('person', 42)
    expect(err.exitCode).toBe(3)
    expect(err.message).toBe('person not found: 42')
    expect(err.name).toBe('NotFoundError')
  })

  it('should accept string IDs', () => {
    const err = new NotFoundError('person', 'jane-smith')
    expect(err.message).toBe('person not found: jane-smith')
  })
})

describe('ConflictError', () => {
  it('should have exit code 4', () => {
    const err = new ConflictError('duplicate email')
    expect(err.exitCode).toBe(4)
    expect(err.name).toBe('ConflictError')
  })
})

describe('DatabaseError', () => {
  it('should have exit code 10', () => {
    const err = new DatabaseError('database locked')
    expect(err.exitCode).toBe(10)
    expect(err.name).toBe('DatabaseError')
  })

  it('should preserve cause', () => {
    const cause = new Error('SQLITE_BUSY')
    const err = new DatabaseError('database locked', cause)
    expect(err.cause).toBe(cause)
  })
})
