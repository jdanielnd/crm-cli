#!/usr/bin/env node

import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { Command } from 'commander'

import { CliError } from '../models/errors.js'
import { registerContextCommands } from './commands/context.js'
import { registerDealCommands } from './commands/deal.js'
import { registerLogCommands } from './commands/log.js'
import { registerMcpCommands } from './commands/mcp.js'
import { registerOrgCommands } from './commands/org.js'
import { registerPersonCommands } from './commands/person.js'
import { registerRelateCommands } from './commands/relate.js'
import { registerSearchCommands } from './commands/search.js'
import { registerStatusCommands } from './commands/status.js'
import { registerTagCommands } from './commands/tag.js'
import { registerTaskCommands } from './commands/task.js'

const __dirname = dirname(fileURLToPath(import.meta.url))
const pkg = JSON.parse(readFileSync(join(__dirname, '../../package.json'), 'utf-8')) as {
  version: string
}

const program = new Command()

program
  .name('crm')
  .description('Local-first personal CRM for the terminal')
  .version(pkg.version)
  .option('-f, --format <format>', 'output format: table, json, csv, tsv')
  .option('-q, --quiet', 'minimal output (just IDs)')
  .option('-v, --verbose', 'verbose output')
  .option('--db <path>', 'alternate database path')
  .option('--no-color', 'disable colors')

registerPersonCommands(program)
registerOrgCommands(program)
registerLogCommands(program)
registerTagCommands(program)
registerRelateCommands(program)
registerDealCommands(program)
registerTaskCommands(program)
registerSearchCommands(program)
registerContextCommands(program)
registerStatusCommands(program)
registerMcpCommands(program)

async function main() {
  try {
    await program.parseAsync(process.argv)
  } catch (error: unknown) {
    if (error instanceof CliError) {
      process.stderr.write(`crm: error: ${error.message}\n`)
      if (program.opts()['verbose'] && error.stack) {
        process.stderr.write(`${error.stack}\n`)
      }
      process.exitCode = error.exitCode
    } else if (error instanceof Error) {
      process.stderr.write(`crm: error: ${error.message}\n`)
      if (program.opts()['verbose'] && error.stack) {
        process.stderr.write(`${error.stack}\n`)
      }
      process.exitCode = 1
    }
  }
}

void main()
