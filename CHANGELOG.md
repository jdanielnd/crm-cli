# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.6] - 2026-03-08

### Added

- Project website at [crmcli.sh](https://www.crmcli.sh) with install guide, FAQ, and AI use case examples
- `AGENT.md` with CRM agent instructions for Claude users
- Claude Desktop setup instructions in README

### Changed

- Improved code quality across CLI commands, repositories, formatters, and tests
- Improved README with banner, badges, and clearer structure

### Fixed

- Hardened input validation, error sanitization, and integer overflow checks in MCP server
- Restricted `~/.crm/` directory permissions to user-only (0700)
- CSV headers now use title-case to match table output
- Mobile responsiveness fixes for website (FAQ cards, MCP section, footer)

### Security

- MCP server no longer leaks internal error messages to clients
- Added interaction type and direction validation
- Added entity existence checks before applying tags

## [0.1.5] - 2026-03-07

### Added

- Duplicate person detection on create (by email)
- `crm_person_delete` MCP tool

## [0.1.4] - 2026-03-07

### Fixed

- Code review findings across codebase

## [0.1.3] - 2026-03-07

### Added

- `crm_person_relate` MCP tool for creating relationships between people via AI agents

## [0.1.2] - 2026-03-07

### Changed

- Switched Homebrew distribution from cask to formula
- Rewrote README to document actual features

## [0.1.1] - 2026-03-07

### Added

- Person CRUD commands (`add`, `list`, `show`, `edit`, `delete`)
- Organization CRUD commands
- Interaction logging (`call`, `email`, `meeting`, `note`, `message`)
- Polymorphic tagging system
- Person-to-person relationships (`colleague`, `friend`, `manager`, `mentor`, `referred-by`)
- Deals and pipeline management with stages
- Tasks and follow-ups with priorities and due dates
- Cross-entity full-text search (FTS5)
- Context briefing command for pre-meeting prep
- Dashboard summary (`crm status`)
- MCP server with 17 tools for AI agent integration
- Output formats: table, JSON, CSV, TSV
- Quiet mode (`-q`) for scripting
- GitHub Actions CI/CD
- GoReleaser for cross-platform builds
- Homebrew tap

[0.1.6]: https://github.com/jdanielnd/crm-cli/releases/tag/v0.1.6
[0.1.5]: https://github.com/jdanielnd/crm-cli/releases/tag/v0.1.5
[0.1.4]: https://github.com/jdanielnd/crm-cli/releases/tag/v0.1.4
[0.1.3]: https://github.com/jdanielnd/crm-cli/releases/tag/v0.1.3
[0.1.2]: https://github.com/jdanielnd/crm-cli/releases/tag/v0.1.2
[0.1.1]: https://github.com/jdanielnd/crm-cli/releases/tag/v0.1.1
