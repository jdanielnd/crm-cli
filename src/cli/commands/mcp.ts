import { Command } from 'commander'

interface GlobalOpts {
  db?: string
}

function getOpts(cmd: Command): GlobalOpts {
  let root = cmd
  while (root.parent) root = root.parent
  return root.opts()
}

export function registerMcpCommands(program: Command): void {
  const mcp = program.command('mcp').description('MCP server for AI agents')

  mcp
    .command('serve')
    .description('Start MCP server (stdio transport)')
    .action(async (_opts: Record<string, unknown>, cmd: Command) => {
      const globalOpts = getOpts(cmd)
      const { startMcpServer } = await import('../../mcp/server.js')
      await startMcpServer(globalOpts.db)
    })
}
