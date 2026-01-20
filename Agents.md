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

## Project Structure

```
asc/
├── main.go                    # Entry point
├── cmd/
│   ├── commands.go           # Core commands (feedback, crashes, reviews)
│   └── auth.go               # Authentication commands
├── internal/
│   ├── asc/                  # ASC API client
│   ├── auth/                 # Credential handling (config)
│   └── config/               # Configuration management
├── Makefile                  # Build commands
└── .github/workflows/        # CI/CD
```

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

All commands support `--json` for easy parsing by AI agents. JSON output is **minified** (one line) to minimize token usage.

Output formats:
| Format | Flag | Use Case |
|--------|------|----------|
| JSON (minified) | `--json` | AI agents, scripting |
| Table | `--output table` | Humans in terminal |
| Markdown | `--output markdown` | Humans, documentation |

## Commands

### Core Commands (v1)

```bash
# TestFlight - JSON for AI agents
asc feedback --app "123456789" --json
asc crashes --app "123456789" --json

# TestFlight - Table for humans
asc feedback --app "123456789" --output table

# TestFlight - Markdown for docs
asc crashes --app "123456789" --output markdown

# App Store - JSON for AI agents
asc reviews --app "123456789" --json

# App Store - Table for humans
asc reviews --app "123456789" --stars 1 --output table

# Authentication
asc auth login --name "MyKey" --key-id "ABC" --issuer-id "DEF" --private-key /path/to/key.p8
asc auth status
```

### Future Commands (v2+)

- `asc localizations upload/download`
- `asc submit` - Ship builds
- `asc sandbox` - Create test users
- `asc apps` - List apps
- `asc builds` - Manage builds

## Authentication

Uses App Store Connect API keys (not Apple ID). Keys are:
1. Generated at https://appstoreconnect.apple.com/access/integrations/api
2. Stored in the system keychain (with local config fallback)
3. Never committed to version control

Environment variables (fallback):
- `ASC_KEY_ID`
- `ASC_ISSUER_ID`
- `ASC_PRIVATE_KEY_PATH`

## Code Style

- Use `ffcli` for command structure
- Return explicit errors with context
- Support `--json` flag on all commands
- Use Go's standard library where possible
- Write tests for all new functionality

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

## Tips for Claude Code

1. Always run `make test` before committing
2. Use explicit flag names, not short aliases
3. Return JSON-friendly output for AI consumption
4. Don't add interactive prompts - use flags instead
5. Keep commands focused and simple
