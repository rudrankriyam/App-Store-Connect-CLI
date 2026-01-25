---
name: asc-cli-usage
description: Guidance for using the App Store Connect CLI in this repo (flags, output formats, pagination, auth, and discovery). Use when asked to run or design asc commands or interact with App Store Connect via the CLI.
---

# ASC CLI usage

Use this skill when you need to run or design `asc` commands for App Store Connect.

## Command discovery
- Always use `--help` to discover commands and flags.
  - `asc --help`
  - `asc builds --help`
  - `asc builds list --help`

## Flag conventions
- Use explicit long flags (e.g., `--app`, `--output`).
- No interactive prompts; destructive operations require `--confirm`.
- Use `--paginate` when the user wants all pages.

## Output formats
- Default output is minified JSON.
- Use `--output table` or `--output markdown` only for human-readable output.
- `--pretty` is only valid with JSON output.

## Authentication and defaults
- Prefer keychain auth via `asc auth login`.
- Fallback env vars: `ASC_KEY_ID`, `ASC_ISSUER_ID`, `ASC_PRIVATE_KEY_PATH`.
- `ASC_APP_ID` can provide a default app ID.

## Timeouts
- `ASC_TIMEOUT` / `ASC_TIMEOUT_SECONDS` control request timeouts.
- `ASC_UPLOAD_TIMEOUT` / `ASC_UPLOAD_TIMEOUT_SECONDS` control upload timeouts.

## References
- `AGENTS.md`
- `docs/TESTING.md`
- `docs/GO_STANDARDS.md`
- `docs/API_NOTES.md`
- `CONTRIBUTING.md`
