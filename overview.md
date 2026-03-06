# crm — Product Overview

A local-first personal CRM for the terminal. Store contacts, organizations, interactions, deals, and tasks in a local SQLite database. Every command outputs structured data (table, JSON, CSV) making it equally usable by humans and AI agents. Ships with a built-in MCP server so AI agents can search, query, and update your CRM directly.

## Core Principles

- **Local-first** — SQLite in `~/.crm/`, no cloud, no accounts, your data stays on your machine
- **AI-native** — MCP server, JSON output, `crm context` command for pre-meeting briefings
- **Unix philosophy** — pipeable, composable, scriptable, works with `jq`, `fzf`, `xargs`
- **Zero config** — auto-creates database on first use

## Who It's For

- Developers who network and want CLI-native tooling
- Freelancers/consultants managing client relationships
- Founders doing early-stage sales
- Anyone who wants AI agents to help manage their professional network
- Power users who prefer terminal over web UIs

## What Makes It Different

1. **AI-native from day one** — not a web CRM with an API bolted on; MCP server and structured output are core
2. **Privacy-first** — local SQLite, no cloud, no telemetry, portable file you own
3. **Unix philosophy** — pipes, JSON, exit codes, composable with the entire shell ecosystem
4. **Interaction-centric** — not just an address book; centers on the history of your relationships
5. **Zero friction** — `crm person add "Jane Smith"` and you're running

---

## Data Model

Six core entities with full-text search across all of them:

**People** — contacts with name, email, phone, title, location, notes, an AI-maintained `summary` field (living dossier of key facts about the person), plus custom fields

**Organizations** — companies/groups that people belong to

**Interactions** — the activity log: calls, emails, meetings, notes, messages (with timestamps, direction, content)

**Deals** — opportunities with value, stage (lead -> prospect -> proposal -> negotiation -> won/lost), linked to people/orgs

**Tasks** — follow-ups with due dates and priorities, tied to contacts or deals

**Tags** — flat labels applicable to any entity (polymorphic tagging)

### Supporting Structures

- **Custom fields** — key/value pairs on any entity for extensibility (birthday, github handle, etc.)
- **Relationships** — person-to-person links (colleague, friend, manager, mentor, referred-by)
- **Full-text search** — FTS5 indexes on people, organizations, interactions, and deals

### Key Design Decisions

- Integer IDs for human-friendly CLI use (`crm person show 42`) + UUIDs for stable external references
- `occurred_at` vs `created_at` on interactions (log a call after the fact)
- Soft-delete via `archived` flag (CRM data is precious)
- WAL mode for concurrent access (MCP server + CLI simultaneously)
- **AI-maintained `summary` field** on People and Organizations — a living dossier that AI agents update after each interaction. Contains key facts, preferences, relationship context, and conversation history highlights. Separate from `notes` (which is user-written). AI can read the summary before a meeting and update it after, keeping a continuously refined profile of each contact.

---

## CLI Interface

Pattern: `crm <entity> <action> [args] [flags]`

### Global Flags

```
--format, -f    Output format: table (default), json, csv, tsv
--quiet, -q     Minimal output (just IDs, for piping)
--verbose, -v   Verbose output
--db <path>     Alternate database path
--no-color      Disable colors
```

### People

```bash
crm person add "Jane Smith" --email jane@example.com --org "Acme Corp" --tag client
crm person list --tag client --sort last-contacted
crm person show 42 --with interactions
crm person edit 42 --email new@email.com --set birthday=1990-03-15
crm person delete 42                    # soft-delete
crm person merge 42 43                  # merge duplicates
crm person relate 42 43 --type colleague
```

### Organizations

```bash
crm org add "Acme Corp" --domain acme.com --industry SaaS
crm org list
crm org show 5 --with people
```

### Interactions

```bash
crm log call 42 --subject "Discussed roadmap"
crm log email 42 --subject "Follow-up" --direction outbound
crm log meeting 42 43 --subject "Product demo" --at "2026-03-05 14:00"
crm log note 42                          # opens $EDITOR
echo "Call notes" | crm log note 42 --subject "Quick check-in"
```

### Tags

```bash
crm tag list
crm tag apply person 42 "vip"
crm tag remove person 42 "vip"
```

### Deals

```bash
crm deal add "Website Redesign" --value 15000 --person 42 --stage proposal
crm deal list --stage proposal,negotiation
crm deal edit 10 --stage won --closed-at today
crm deal pipeline                        # summary by stage
```

### Tasks

```bash
crm task add "Follow up on proposal" --person 42 --due "next friday" --priority high
crm task list --overdue
crm task done 15
```

### Search

```bash
crm search "jane"                        # searches people, orgs, interactions
crm search "roadmap" --type interaction --since 2026-01-01
```

### Context (AI Killer Feature)

```bash
crm context "Jane Smith"                 # everything about a person in one call
```

Returns: person profile + org + recent interactions + open deals + open tasks + relationships + custom fields + interaction frequency summary.

### Import/Export

```bash
crm import contacts.csv --type person --map "Name=first_name,Email=email"
crm import contacts.vcf                  # vCard
crm export people --format csv > contacts.csv
crm export all --format json > backup.json
```

### Status Dashboard

```bash
crm status
# 342 contacts | 28 organizations | 5 open deals ($47,500)
# 3 tasks overdue | 12 interactions this week
```

---

## Piping & Composition

```bash
# Bulk tag everyone at an org
crm person list --org "Acme Corp" -q | xargs -I{} crm tag apply person {} "acme-team"

# Pipe email content into a log entry
pbpaste | crm log email 42 --subject "Re: Proposal" --direction inbound

# Chain with jq
crm person list -f json | jq '.[] | select(.tags | contains(["vip"])) | .email'

# Interactive selection with fzf
crm person list -f tsv | fzf | cut -f1 | xargs crm person show

# Overdue task notifications
crm task list --overdue -f json | jq -r '.[] | "OVERDUE: \(.title) — \(.person.first_name)"'
```

### Exit Codes

| Code | Meaning    |
| ---- | ---------- |
| 0    | Success    |
| 1    | Error      |
| 2    | Usage error|
| 3    | Not found  |
| 4    | Conflict   |
| 10   | DB error   |

---

## AI Integration (MCP Server)

```bash
crm mcp serve          # stdio transport (for Claude Code, etc.)
```

### MCP Tools

| Tool                        | Description                                           |
| --------------------------- | ----------------------------------------------------- |
| `crm_person_search`         | Search people by name, email, tag, org                |
| `crm_person_get`            | Get full details for a person                         |
| `crm_person_create`         | Create a new person                                   |
| `crm_person_update`         | Update person fields                                  |
| `crm_person_update_summary` | Update the AI-maintained summary/dossier for a person |
| `crm_org_search`            | Search organizations                                  |
| `crm_org_get`               | Get org details with people                           |
| `crm_interaction_log`       | Log an interaction                                    |
| `crm_interaction_list`      | List interactions for a person/org                    |
| `crm_search`                | Cross-entity full-text search                         |
| `crm_context`               | Full context for a person (the briefing)              |
| `crm_task_create`           | Create a follow-up task                               |
| `crm_task_list`             | List open tasks                                       |
| `crm_deal_create`           | Create a deal                                         |
| `crm_deal_update`           | Update deal stage                                     |
| `crm_tag_apply`             | Apply tag to entity                                   |
| `crm_person_relate`         | Create relationship between two people                |
| `crm_stats`                 | CRM summary stats                                     |

### AI Workflow Example — Post-Meeting

```
User: "Just had a great meeting with Jane. We agreed on the timeline, she'll sign next week."

AI calls:
1. crm_interaction_log(type="meeting", person_ids=[42], subject="Timeline agreement")
2. crm_deal_update(id=10, stage="negotiation")
3. crm_task_create(title="Follow up on signed contract", person_id=42, due="next week")
4. crm_person_update_summary(id=42, summary="CTO at Acme Corp. Working on $15K website
   redesign deal. Agreed on timeline Mar 2026, expecting signed contract by mid-March.
   Prefers async communication. Met at ReactConf 2025. Colleague of Tom Baker.")
```

The `summary` field acts as a living dossier — AI reads it before meetings for instant context, and updates it after interactions to capture new facts, preferences, and relationship evolution. Unlike `notes` (user-written), `summary` is AI-maintained and continuously refined.

---

## Data Directory

```
~/.crm/
  crm.db              # SQLite database (WAL mode)
  backups/            # auto-backup before migrations
```

---

## Implementation Phases

**Phase 1 — Foundation (MVP):** Database schema + migrations, Person CRUD, Organization CRUD, basic interaction logging (`crm log`), full-text search, table + JSON output

**Phase 2 — Relationships & Context:** Tags, custom fields, person-to-person relationships, `crm context` command, CSV/vCard import/export

**Phase 3 — AI Integration:** MCP server with full tool suite, `crm mcp serve`, natural language date parsing

**Phase 4 — Productivity:** Deals/pipeline, tasks/follow-ups, `crm status` dashboard, auto-backups

**Phase 5 — Polish:** fzf integration, duplicate detection/merge, bulk operations, shell completions, man page
