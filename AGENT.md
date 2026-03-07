# CRM Agent Instructions

> Copy this file into your project as `CLAUDE.md`, or paste its contents into your Claude project instructions, so Claude knows how to use your CRM effectively.

## Setup

You have access to a local CRM via MCP tools (prefixed with `crm_`). The CRM stores contacts, organizations, interactions, deals, tasks, tags, and relationships in a local SQLite database. Use these tools proactively to keep the CRM up to date.

## Core Workflows

### When the user mentions meeting or talking to someone

1. Search for the person with `crm_person_search` to check if they already exist
2. If they don't exist, create them with `crm_person_create` — capture name, email, title, company, and any other details mentioned
3. Log the interaction with `crm_interaction_log` — use the right type (call, email, meeting, note, message), include a subject summarizing what happened, and add content with key details
4. Update the person's `summary` field with `crm_person_update` — append new facts, don't overwrite existing context
5. If follow-ups are needed, create tasks with `crm_task_create` with realistic due dates

### When the user asks about someone before a meeting

1. Use `crm_context` with the person's ID — this returns their full profile, organization, recent interactions, open deals, pending tasks, relationships, and tags in one call
2. Summarize the key points: who they are, last interaction, any open deals or pending tasks, and what was discussed previously
3. Flag any overdue tasks or stale follow-ups

### When the user mentions a deal or opportunity

1. Check if a deal already exists with `crm_search`
2. Create or update the deal with the right stage: lead → prospect → proposal → negotiation → won/lost
3. Link the deal to the right person and organization
4. If a deal moves to "won" or "lost", set the `closed_at` date

### When the user meets multiple people at an event

1. Create all contacts, capturing as much detail as mentioned
2. Tag everyone with the event name (e.g., "reactconf-2026") using `crm_tag_apply`
3. Log a note interaction for each person with context about the conversation
4. Create relationships between people if mentioned (e.g., "Sarah introduced me to David")
5. Set follow-up tasks with staggered due dates so the user isn't overwhelmed

## The Summary Field

Each person has a `summary` field. Treat it as a **living dossier** that you maintain:

- Read it before any conversation about that person for context
- After interactions, update it with new facts: preferences, interests, what was discussed, decisions made, personal details they shared
- Keep it concise and structured — bullet points work well
- Never overwrite — append and consolidate
- Include things like: communication preferences, timezone, key interests, relationship dynamics, important dates mentioned

Example summary:
```
- CTO at Acme Corp, reports to CEO Maria Lopez
- Focused on developer experience and reducing build times
- Prefers async communication (Slack/email over calls)
- Based in SF, has two kids, coaches youth soccer
- Met at ReactConf 2026, bonded over Rust adoption
- Interested in our API product, concerned about pricing
- Decision timeline: Q2 2026 budget cycle
```

## Tags

Use tags liberally to organize contacts:

- Event tags: "reactconf-2026", "yc-demo-day"
- Relationship tags: "investor", "customer", "partner", "mentor"
- Status tags: "vip", "churned", "warm-intro"
- Project tags: "project-alpha", "series-a"

Tags work on any entity: person, organization, deal, interaction.

## Interactions

Always log interactions with the right type and direction:

- **call** — phone/video calls (direction: inbound or outbound)
- **email** — email exchanges (direction: inbound or outbound)
- **meeting** — in-person or scheduled video meetings
- **note** — observations, internal notes, things to remember
- **message** — chat messages (Slack, iMessage, WhatsApp, etc.)

Link interactions to all relevant people using the `person_ids` array.

## Tasks

When creating follow-ups:

- Set realistic due dates based on what was discussed
- Use priority levels meaningfully: **high** for time-sensitive commitments, **medium** for standard follow-ups, **low** for nice-to-haves
- Link tasks to the relevant person and deal when applicable
- Check `crm_task_list` with `overdue: true` periodically and remind the user

## Relationships

Connect people to each other when the user mentions how they know each other:

- **colleague** — work together
- **friend** — personal relationship
- **manager** — reports-to relationship
- **mentor** — mentorship
- **referred-by** — one person introduced the other

Always add notes explaining the connection context.

## General Principles

- **Be proactive**: if the user mentions someone, check the CRM. Don't wait to be asked.
- **Capture everything**: names, companies, emails, context — anything mentioned is worth storing.
- **Keep it current**: update summaries and deal stages after every relevant conversation.
- **Connect the dots**: link people to orgs, deals to people, tag related contacts.
- **Remind naturally**: if you notice overdue tasks or stale deals while looking up a contact, mention them.
- **Don't duplicate**: always search before creating to avoid duplicate contacts.
