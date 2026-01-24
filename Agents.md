# CLAUDE.md

A fast, lightweight, AI-agent-friendly CLI for App Store Connect. Built in Go with [ffcli](https://github.com/peterbourgon/ff).

## Core Principles

- **Explicit flags**: Always `--app` not `-a`, `--output` not `-o`
- **JSON-first**: Minified JSON by default (saves tokens), `--output table/markdown` for humans
- **No interactive prompts**: Use `--confirm` flags for destructive operations
- **Pagination**: `--paginate` fetches all pages automatically

## Discovering Commands

**Use `--help` to discover commands and flags.** The CLI is self-documenting:

```bash
asc --help                    # List all commands
asc builds --help             # List builds subcommands
asc builds list --help        # Show all flags for a command
```

Do not memorize commands. Always check `--help` for the current interface.

## Build & Test

```bash
make build      # Build binary
make test       # Run tests (always run before committing)
make lint       # Lint code
make format     # Format code
```

## Authentication

API keys are generated at https://appstoreconnect.apple.com/access/integrations/api and stored in the system keychain (with local config fallback). Never commit keys to version control.

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `ASC_KEY_ID`, `ASC_ISSUER_ID`, `ASC_PRIVATE_KEY_PATH` | Auth fallback |
| `ASC_APP_ID` | Default app ID |
| `ASC_VENDOR_NUMBER` | Sales/finance reports |
| `ASC_TIMEOUT` | Request timeout (e.g., `90s`, `2m`) |
| `ASC_TIMEOUT_SECONDS` | Timeout in seconds (alternative) |
| `ASC_UPLOAD_TIMEOUT` | Upload timeout (e.g., `60s`, `2m`) |
| `ASC_UPLOAD_TIMEOUT_SECONDS` | Upload timeout in seconds (alternative) |

## Further Reading

Detailed guidance on specific topics (only read when needed):

- **Go coding standards**: `docs/GO_STANDARDS.md`
- **Testing patterns**: `docs/TESTING.md`
- **Git workflow, adding features**: `docs/CONTRIBUTING.md`
- **API quirks (analytics, finance, sandbox)**: `docs/API_NOTES.md`
- **Development setup, PRs**: `CONTRIBUTING.md` (root)
