#!/usr/bin/env node

import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { Command } from 'commander'

import { CliError } from '../models/errors.js'

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

// Commands will be registered here in subsequent steps

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
