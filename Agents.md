# CLAUDE.md

This file provides guidance to Claude Code when working with this project.

## Project Overview

**ASC** (App Store Connect CLI) is a fast, lightweight, AI-agent-friendly CLI for App Store Connect. Built in Go, it enables developers and AI agents to ship iOS apps with zero friction.

## Core Values

1. **Speed** - Fast startup, fast execution
2. **Simplicity** - Minimal config, no plugins, just commands
3. **Explicit over Cryptic** - `--app` not `-a`, `--stars` not `-s`
4. **AI-First** - JSON output by default, clean exit codes, no interactive prompts
5. **Security** - Credentials stored in the system keychain when available

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: [ffcli](https://github.com/peterbourgon/ff) (no globals, functional style)
- **Testing**: Go's built-in testing
- **Distribution**: Homebrew

## Key Design Decisions

### ffcli over Cobra

We use ffcli because:
- No global state
- Functional composition
- Easier to test
- Cleaner architecture

### Explicit Flags

Always use long-form flags with clear names:
- ✅ `--email`, `--app`, `--output`
- ❌ `-e`, `-a`, `-o`

### JSON-First Output

All commands output minified JSON by default to minimize token usage.

Output formats:
| Format | Flag | Use Case |
|--------|------|----------|
| JSON (minified) | default | AI agents, scripting |
| Table | `--output table` | Humans in terminal |
| Markdown | `--output markdown` | Humans, documentation |

## Commands

### Core Commands (v1)

```bash
# TestFlight - JSON for AI agents
asc feedback --app "123456789"
asc crashes --app "123456789"

# TestFlight - Table for humans
asc feedback --app "123456789" --output table

# TestFlight - Markdown for docs
asc crashes --app "123456789" --output markdown

# App Store - JSON for AI agents
asc reviews --app "123456789"

# App Store - Table for humans
asc reviews --app "123456789" --stars 1 --output table

# Apps & Builds - JSON for AI agents
asc apps
asc apps --sort name
asc builds --app "123456789"
asc builds --app "123456789" --sort -uploadedDate
asc builds info --build "BUILD_ID"
asc builds expire --build "BUILD_ID"

# Utilities
asc version

# Authentication
asc auth login --name "MyKey" --key-id "ABC" --issuer-id "DEF" --private-key /path/to/key.p8
asc auth status
```

### Future Commands (v2+)

- `asc localizations upload/download`
- `asc submit` - Ship builds
- `asc sandbox` - Create test users

## Authentication

Uses App Store Connect API keys (not Apple ID). Keys are:
1. Generated at https://appstoreconnect.apple.com/access/integrations/api
2. Stored in the system keychain (with local config fallback)
3. Never committed to version control

Environment variables (fallback):
- `ASC_KEY_ID`
- `ASC_ISSUER_ID`
- `ASC_PRIVATE_KEY_PATH`

Analytics & sales env:
- `ASC_VENDOR_NUMBER`
- `ASC_TIMEOUT` (duration like `90s`, `2m`)
- `ASC_TIMEOUT_SECONDS`

## Analytics & Sales Notes

- Sales report date formats: DAILY/WEEKLY `YYYY-MM-DD`, MONTHLY `YYYY-MM`, YEARLY `YYYY`
- Vendor number comes from Sales and Trends → Reports URL (`vendorNumber=...`)
- Use `--paginate` with `asc analytics get --date` to avoid missing instances on later pages
- Long analytics runs can require raising `ASC_TIMEOUT`

## Code Style

- Use `ffcli` for command structure
- Return explicit errors with context
- Output JSON by default; use `--output` for table/markdown
- Use Go's standard library where possible
- Write tests for all new functionality

## Go Standards

Follow idiomatic Go so the code is predictable to anyone who reads Go:

- **Formatting:** always run `gofmt` (and `gofumpt` via `make format`). No manual formatting.
- **Naming:** use mixedCaps; keep common initialisms uppercase (`ID`, `URL`, `API`, `JSON`).
- **Errors:** return errors, don’t panic for expected failures. Wrap with context using `%w`.
- **Context:** pass `context.Context` into network operations; respect timeouts and cancellations.
- **Types:** model request/response types with JSON tags; use pointers for optional fields, values for required fields.
- **Enums:** prefer typed `const` values (not raw strings) for API enums and resource types.
- **CLI behavior:** if a flag is accepted, it must be implemented or error; never silently ignore flags.
- **Output:** data goes to stdout, errors to stderr; keep JSON minified by default.
- **Dependencies:** standard library first; avoid new deps unless necessary and justified.
- **Tests:** deterministic, table‑driven when possible; use `t.Helper()`. For JSON, unmarshal and assert fields (not `strings.Contains`). Cover success + validation + API error paths.

## CLI Help Output

- Use `UsageFunc` on ffcli commands for consistent help formatting (bold headers, aligned flags).
- When returning `flag.ErrHelp`, do **not** call `fs.Usage()` manually, or help output will be duplicated.
- Help output is written to stderr by default; tests should capture stderr for usage text.

## Building

```bash
make build      # Build binary
make test       # Run tests
make lint       # Lint code
make format     # Format code
make install    # Install locally
```

## Testing Guidelines

- Write tests for all exported functions
- Use table-driven tests
- Mock external API calls
- Test error cases
- Add CLI-level tests in `cmd/commands_test.go` for command output/parsing
- Prefer test-driven development (write tests first, then implement)
- Cover success, validation, and API error paths for each client endpoint

## Common Tasks

### Adding a New Command

1. Add a factory in `cmd/commands.go` or a new `cmd/*.go`
2. Use ffcli pattern from existing commands
3. Add to `RootCommand` subcommands list
4. Write tests

### Adding a New API Endpoint

1. Add method to `internal/asc/client.go`
2. Add types for request/response
3. Add helper functions for output
4. Add command in `cmd/` to use it

## Releases

- Tag releases with plain semver like `0.1.0` (no `v` prefix).

## Git Workflow

- Branch from `main` and keep one logical change per branch
- Do not commit directly to `main` unless explicitly instructed; prefer PRs
- Prefer `git worktree add` for parallel tasks; remove with `git worktree remove` when done
- Keep worktrees clean: run `git status` before/after changes
- Rebase on `main` before merging; avoid merge commits
- Commit small, coherent changes; no WIP commits on shared branches
- Use concise, present-tense commit messages that match repo style
- Review `git diff` before staging; stage only what you intend
- Never commit secrets or local config files (keys, `.env`, `config.json`)
- Run `make format`, `make lint`, and `make test` before committing code changes
- Avoid rewriting shared history or force pushes unless explicitly required

## Tips for Claude Code

1. Always run `make test` before committing
2. Use explicit flag names, not short aliases
3. Return JSON-friendly output for AI consumption
4. Don't add interactive prompts - use flags instead
5. Keep commands focused and simple
6. When responding to audit feedback, prefer `codex exec` to implement fixes and search the internet for missing details; if `codex exec` isn't available, proceed manually and note the limitation.
