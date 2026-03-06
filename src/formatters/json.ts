export function formatJson(data: unknown): string {
  return JSON.stringify(data, null, process.stdout.isTTY ? 2 : 0)
}
