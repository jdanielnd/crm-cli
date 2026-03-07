# crm — your contacts, in your terminal.

![crm banner](.github/banner-crm-cli.png)

[![CI](https://github.com/jdanielnd/crm-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/jdanielnd/crm-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdanielnd/crm-cli)](https://goreportcard.com/report/github.com/jdanielnd/crm-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A local-first personal CRM that lives in your terminal. Manage contacts, organizations, interactions, deals, and tasks — all from the command line. Ships with a built-in [MCP server](https://modelcontextprotocol.io/) so AI agents like Claude can read and update your CRM directly.

Single static binary. SQLite database. No cloud. No accounts. Your data stays on your machine.

## Features

- **People & Organizations** — contacts with full profiles, linked to companies
- **Interaction Log** — track calls, emails, meetings, notes, and messages
- **Deals & Pipeline** — sales opportunities with stages, values, and a pipeline view
- **Tasks** — follow-ups with due dates, priorities, and completion tracking
- **Tags** — labels on any entity for flexible organization
- **Relationships** — person-to-person links (colleague, mentor, referred-by, …)
- **Full-Text Search** — find anything across all entities instantly
- **Context Briefing** — one command to get everything about a person before a meeting
- **AI-Ready** — built-in MCP server for Claude and other AI agents
- **Multiple Output Formats** — table, JSON, CSV, TSV — pipe into anything
- **Zero Dependencies** — single binary, pure-Go SQLite, runs anywhere

## Install

```bash
# Homebrew (macOS/Linux)
brew install jdanielnd/tap/crm

# Go
go install github.com/jdanielnd/crm-cli/cmd/crm@latest

# Or download a binary from GitHub Releases
# https://github.com/jdanielnd/crm-cli/releases
```

## Quick Start

```bash
# Add a company
crm org add "Acme Corp" --domain acme.com

# Add a contact and link them to the org
crm person add "Jane Smith" --email jane@example.com --org 1

# Log a call
crm log call 1 --subject "Intro call"

# Get a full briefing before your next meeting
crm context 1

# See your dashboard
crm status
```

## Commands

### People

```bash
crm person add "Jane Smith" --email jane@example.com --phone 555-1234 --org 1
crm person list                          # all contacts
crm person list --tag vip --limit 10     # filtered
crm person show 1                        # full details
crm person edit 1 --title "CTO" --company "Acme Corp"
crm person delete 1                      # soft-delete (recoverable)
```

### Organizations

```bash
crm org add "Acme Corp" --domain acme.com --industry SaaS
crm org list
crm org show 1 --with-people             # show org + its members
crm org edit 1 --domain acme.io
crm org delete 1
```

### Interactions

```bash
crm log call 1 --subject "Discussed roadmap" --content "Agreed on Q2 timeline"
crm log email 1 --subject "Follow-up" --direction outbound
crm log meeting 1 2 --subject "Product demo" --at "2026-03-05 14:00"
crm log note 1 --subject "Quick check-in" --content "Seemed interested"
crm log list --person 1 --limit 5        # recent interactions with person
```

### Deals

```bash
crm deal add "Website Redesign" --value 15000 --person 1 --stage proposal
crm deal list --open                     # exclude won/lost
crm deal list --stage proposal
crm deal show 1
crm deal edit 1 --stage won --closed-at 2026-03-15
crm deal delete 1
crm deal pipeline                        # summary by stage
```

### Tasks

```bash
crm task add "Follow up on proposal" --person 1 --due 2026-03-14 --priority high
crm task list                            # open tasks
crm task list --overdue                  # past due
crm task list --all                      # include completed
crm task show 1
crm task edit 1 --priority medium
crm task done 1                          # mark completed
crm task delete 1
```

### Tags

```bash
crm tag apply person 1 "vip"
crm tag apply deal 1 "q2"
crm tag show person 1                    # tags on a person
crm tag remove person 1 "vip"
crm tag list                             # all tags
crm tag delete "old-tag"                 # remove tag entirely
```

### Relationships

```bash
crm person relate 1 2 --type colleague --notes "Met at ReactConf"
crm person relationships 1               # list all relationships
crm person unrelate 1                    # remove by relationship ID
```

Relationship types: `colleague`, `friend`, `manager`, `mentor`, `referred-by`

### Search

```bash
crm search "jane"                        # searches people, orgs, interactions, deals
crm search "roadmap" --type interaction  # filter by entity type
```

### Context Briefing

```bash
crm context 1                            # full briefing by person ID
crm context "Jane Smith"                 # or by name
```

Returns person profile, organization, recent interactions, open deals, pending tasks, relationships, and tags — everything you need before a meeting.

### Dashboard

```bash
crm status
```

Shows contact/org counts, open deals with total value, task counts, overdue items, pipeline breakdown, and recent activity.

## Output Formats

Every command supports `--format` / `-f`:

```bash
crm person list                          # table (default in TTY)
crm person list -f json                  # JSON (default when piped)
crm person list -f csv                   # CSV
crm person list -f tsv                   # TSV
```

Use `-q` / `--quiet` for minimal output (just IDs), useful for scripting:

```bash
crm person list --tag vip -q | xargs -I{} crm tag apply person {} "priority"
```

## Piping & Composition

`crm` follows Unix conventions — data to stdout, messages to stderr, structured exit codes — so it plays well with other tools:

```bash
# Bulk tag everyone at an org
crm person list -f json | jq '.[] | select(.org_id == 1) | .id' | xargs -I{} crm tag apply person {} "acme-team"

# Interactive selection with fzf
crm person list -f tsv | fzf | cut -f1 | xargs crm person show

# Export contacts to CSV
crm person list -f csv > contacts.csv

# Overdue task notifications
crm task list --overdue -f json | jq -r '.[] | "OVERDUE: \(.title)"'
```

Exit codes: `0` success, `1` error, `2` usage, `3` not found, `4` conflict, `10` database error.

## AI Integration (MCP Server)

`crm` includes a built-in [MCP](https://modelcontextprotocol.io/) server so AI agents can query and update your CRM over a structured protocol. Add it to your Claude Code or Claude Desktop config:

```json
{
  "mcpServers": {
    "crm": {
      "command": "crm",
      "args": ["mcp", "serve"]
    }
  }
}
```

### Available Tools

| Tool | Description |
|------|-------------|
| `crm_person_search` | Search people by name, email, tag, org |
| `crm_person_get` | Get full details for a person |
| `crm_person_create` | Create a new person |
| `crm_person_update` | Update person fields |
| `crm_person_delete` | Delete (archive) a person |
| `crm_person_relate` | Create relationship between people |
| `crm_org_search` | Search organizations |
| `crm_org_get` | Get org details with members |
| `crm_interaction_log` | Log a call, email, meeting, or note |
| `crm_interaction_list` | List interactions for a person |
| `crm_search` | Cross-entity full-text search |
| `crm_context` | Full person briefing |
| `crm_deal_create` | Create a deal |
| `crm_deal_update` | Update deal stage/fields |
| `crm_task_create` | Create a task |
| `crm_task_list` | List open tasks |
| `crm_tag_apply` | Apply a tag |
| `crm_stats` | CRM summary stats |

### Example AI Workflow

After a meeting, tell Claude:

> "Just had a great meeting with Jane. We agreed on the timeline, she'll sign next week."

Claude can:
1. Log the interaction with `crm_interaction_log`
2. Update the deal stage with `crm_deal_update`
3. Create a follow-up task with `crm_task_create`
4. Update Jane's summary with `crm_person_update`

The `summary` field on contacts acts as a living dossier that AI agents maintain — reading it before meetings for context, updating it after interactions with new facts and preferences.

## Configuration

The database lives at `~/.crm/crm.db` by default. Override with:

```bash
# Flag (highest priority)
crm --db /path/to/crm.db person list

# Environment variable
export CRM_DB=/path/to/crm.db
```

The database and directory are auto-created on first use.

### iCloud Sync (macOS)

Point your database to iCloud Drive for automatic backup across machines:

```bash
export CRM_DB="$HOME/Library/Mobile Documents/com~apple~CloudDocs/crm/crm.db"
```

SQLite WAL mode handles concurrent reads well, but avoid writing from two machines simultaneously.

## Building from Source

```bash
git clone https://github.com/jdanielnd/crm-cli.git
cd crm-cli
go build -o crm ./cmd/crm
go test ./...
```

Requires Go 1.23+. No CGO — builds anywhere Go runs.

## Tech Stack

| Component | Choice |
|-----------|--------|
| Language | Go |
| Database | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| MCP | [mcp-go](https://github.com/mark3labs/mcp-go) |
| Search | SQLite FTS5 |

## License

MIT
