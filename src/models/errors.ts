export class CliError extends Error {
  constructor(
    message: string,
    public readonly exitCode: number = 1,
  ) {
    super(message)
    this.name = 'CliError'
  }
}

export class ValidationError extends CliError {
  constructor(message: string) {
    super(message, 2)
    this.name = 'ValidationError'
  }
}

export class NotFoundError extends CliError {
  constructor(entity: string, id: number | string) {
    super(`${entity} not found: ${String(id)}`, 3)
    this.name = 'NotFoundError'
  }
}

export class ConflictError extends CliError {
  constructor(message: string) {
    super(message, 4)
    this.name = 'ConflictError'
  }
}

export class DatabaseError extends CliError {
  constructor(message: string, cause?: Error) {
    super(message, 10)
    this.name = 'DatabaseError'
    if (cause) this.cause = cause
  }
}
