# ASC - App Store Connect CLI

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Homebrew-compatible-blue?style=for-the-badge" alt="Homebrew">
</p>

A **fast**, **lightweight**, and **AI-agent friendly** CLI for App Store Connect. Ship iOS apps with zero friction.

## Why ASC?

| Problem | Solution |
|---------|----------|
| Manual App Store Connect work | Automate everything from CLI |
| Slow, heavy tooling | Go binary, fast startup |
| Not AI-agent friendly | JSON output, explicit flags, clean exit codes |

## Quick Start

### Install

```bash
# Via Homebrew (recommended)
brew tap rudrankriyam/tap
brew install rudrankriyam/tap/asc

# Install script (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/rudrankriyam/App-Store-Connect-CLI/main/install.sh | bash

# Installs to ~/.local/bin by default (ensure it's on your PATH)

# Or build from source
git clone https://github.com/rudrankriyam/App-Store-Connect-CLI.git
cd App-Store-Connect-CLI
make build
./asc --help
```

### Authenticate

```bash
# Register your App Store Connect API key
asc auth login \
  --name "MyApp" \
  --key-id "ABC123" \
  --issuer-id "DEF456" \
  --private-key /path/to/AuthKey.p8
```

Generate API keys at: https://appstoreconnect.apple.com/access/integrations/api

Credentials are stored in the system keychain when available, with a local config fallback
at `~/.asc/config.json` (restricted permissions).
Environment variable fallback:
- `ASC_KEY_ID`
- `ASC_ISSUER_ID`
- `ASC_PRIVATE_KEY_PATH`

App ID fallback:
- `ASC_APP_ID`

Analytics & sales env:
- `ASC_VENDOR_NUMBER` (Sales, Trends, and Finance reports)
- `ASC_ANALYTICS_VENDOR_NUMBER` (fallback for analytics vendor number)
- `ASC_TIMEOUT` (e.g., `90s`, `2m`)
- `ASC_TIMEOUT_SECONDS` (e.g., `120`)
- `ASC_UPLOAD_TIMEOUT` (e.g., `60s`, `2m`)
- `ASC_UPLOAD_TIMEOUT_SECONDS` (e.g., `120`)

Retry behavior env:
- `ASC_MAX_RETRIES` (default: 3) for GET/HEAD requests
- `ASC_BASE_DELAY` (default: `1s`)
- `ASC_MAX_DELAY` (default: `30s`)
- `ASC_RETRY_LOG=1` to log retries to stderr
- Retry errors include `retry after` in the final error message when available

## Commands

### Agent Quickstart

- JSON output is default for machine parsing; add `--pretty` when debugging.
- Use `--paginate` to automatically fetch all pages (recommended for AI agents).
- `--paginate` works on list commands including apps, builds list, feedback, crashes, reviews, versions list, pre-release versions list, localizations list, build-localizations list, beta-groups list, beta-testers list, sandbox list, analytics requests/get, testflight apps list, and Xcode Cloud workflows/build-runs.
- Use `--limit` + `--next "<links.next>"` for manual pagination control.
- Sort with `--sort` (prefix `-` for descending):
  - Feedback/Crashes: `createdDate` / `-createdDate`
  - Reviews: `rating` / `-rating`, `createdDate` / `-createdDate`
  - Apps: `name` / `-name`, `bundleId` / `-bundleId`
  - Builds: `uploadedDate` / `-uploadedDate`

### TestFlight

```bash
# List beta feedback (JSON - best for AI agents)
asc feedback --app "123456789"

# Filter feedback by device model and OS version
asc feedback --app "123456789" --device-model "iPhone15,3" --os-version "17.2"

# Filter feedback by platform/build/tester
asc feedback --app "123456789" --app-platform IOS --device-platform IOS --build "BUILD_ID" --tester "TESTER_ID"

# Fetch all feedback pages automatically (AI agents)
asc feedback --app "123456789" --paginate

# Get crash reports (table format - for humans)
asc crashes --app "123456789" --output table

# Get crash reports (markdown - for docs)
asc crashes --app "123456789" --output markdown

# Limit results per page (pagination)
asc crashes --app "123456789" --limit 25

# Sort crashes by created date (newest first)
asc crashes --app "123456789" --sort -createdDate --limit 5

# Fetch all crash pages automatically (AI agents)
asc crashes --app "123456789" --paginate

# List TestFlight apps
asc testflight apps list

# Fetch a TestFlight app by ID
asc testflight apps get --app "APP_ID"

# Export TestFlight configuration to YAML
asc testflight sync pull --app "APP_ID" --output "./testflight.yaml"
```

### Beta Groups

```bash
# List beta groups for an app
asc beta-groups list --app "APP_ID"

# Fetch all beta groups (all pages)
asc beta-groups list --app "APP_ID" --paginate

# Create, fetch, update, delete
asc beta-groups create --app "APP_ID" --name "Beta Testers"
asc beta-groups get --id "GROUP_ID"
asc beta-groups update --id "GROUP_ID" --name "New Name"
asc beta-groups delete --id "GROUP_ID" --confirm

# Add/remove testers
asc beta-groups add-testers --group "GROUP_ID" --tester "TESTER_ID"
asc beta-groups remove-testers --group "GROUP_ID" --tester "TESTER_ID"
```

### Beta Testers

```bash
# List beta testers
asc beta-testers list --app "APP_ID"

# Filter by build or group
asc beta-testers list --app "APP_ID" --build "BUILD_ID"
asc beta-testers list --app "APP_ID" --group "Beta"

# Fetch all beta testers (all pages)
asc beta-testers list --app "APP_ID" --paginate

# Get, add, remove, invite
asc beta-testers get --id "TESTER_ID"
asc beta-testers add --app "APP_ID" --email "tester@example.com" --group "Beta"
asc beta-testers remove --app "APP_ID" --email "tester@example.com"
asc beta-testers invite --app "APP_ID" --email "tester@example.com"

# Manage group membership
asc beta-testers add-groups --id "TESTER_ID" --group "GROUP_ID"
asc beta-testers remove-groups --id "TESTER_ID" --group "GROUP_ID"
```

### App Store

```bash
# List customer reviews (JSON - best for AI agents)
asc reviews --app "123456789"

# Filter by stars (table format - for humans)
asc reviews --app "123456789" --stars 1 --output table

# Filter by territory (markdown - for docs)
asc reviews --app "123456789" --territory US --output markdown

# Sort reviews by created date (newest first)
asc reviews --app "123456789" --sort -createdDate --limit 5

# Fetch all reviews pages automatically (AI agents)
asc reviews --app "123456789" --paginate

# Respond to a customer review
asc reviews respond --review-id "REVIEW_ID" --response "Thanks for your feedback!"

# Get a review response by ID
asc reviews response get --id "RESPONSE_ID"

# Get the response for a specific review
asc reviews response for-review --review-id "REVIEW_ID"

# Delete a review response
asc reviews response delete --id "RESPONSE_ID" --confirm
```

### Analytics & Sales

```bash
# Download daily sales summary (writes .tsv.gz)
asc analytics sales --vendor "12345678" --type SALES --subtype SUMMARY --frequency DAILY --date "2024-01-20"

# Download and decompress
asc analytics sales --vendor "12345678" --type SALES --subtype SUMMARY --frequency DAILY --date "2024-01-20" --decompress

# Create analytics report request
asc analytics request --app "123456789" --access-type ONGOING

# List analytics report requests (all pages)
asc analytics requests --app "123456789" --paginate

# Get analytics reports with instances
asc analytics get --request-id "REQUEST_ID"

# Get analytics report instances for a specific date
asc analytics get --request-id "REQUEST_ID" --date "2024-01-20"

# Include report segments in the output
asc analytics get --request-id "REQUEST_ID" --include-segments

# Fetch a specific instance and include segments
asc analytics get --request-id "REQUEST_ID" --instance-id "INSTANCE_ID" --include-segments

# Download analytics report data
asc analytics download --request-id "REQUEST_ID" --instance-id "INSTANCE_ID"
```

Notes:
- Sales report date formats: DAILY/WEEKLY `YYYY-MM-DD`, MONTHLY `YYYY-MM`, YEARLY `YYYY`
- Reports may not be available yet; ASC returns availability errors when data is pending
- Use `ASC_TIMEOUT` or `ASC_TIMEOUT_SECONDS` for long analytics pagination
- `asc analytics get --date ... --paginate` will scan all report pages (slower, but avoids missing instances)

### Finance Reports

```bash
# Download consolidated report (all regions in one file)
asc finance reports --vendor "12345678" --report-type FINANCIAL --region "ZZ" --date "2025-12"

# Download US-only monthly report
asc finance reports --vendor "12345678" --report-type FINANCIAL --region "US" --date "2025-12"

# Download detailed report (transaction-level data) and decompress
asc finance reports --vendor "12345678" --report-type FINANCE_DETAIL --region "Z1" --date "2025-12" --decompress

# List finance report region codes and currencies
asc finance regions --output table
```

**Report Types (API to UI mapping):**

| API Report Type  | UI Option                              | Region Code(s)          |
|------------------|----------------------------------------|-------------------------|
| `FINANCIAL`      | All Countries or Regions (Single File) | `ZZ` (consolidated)     |
| `FINANCIAL`      | All Countries or Regions (Multiple Files) | `US`, `EU`, `JP`, etc. |
| `FINANCE_DETAIL` | All Countries or Regions (Detailed)    | `Z1` (required)         |
| Not available    | Transaction Tax (Single File)          | N/A - manual download   |

**Notes:**
- Report date format: `YYYY-MM` (Apple fiscal month)
- Reports typically appear the first Friday of the following fiscal month
- `FINANCE_DETAIL` requires region code `Z1` (the only valid region for detailed reports)
- Transaction Tax reports are not available via API - download manually from App Store Connect
- Use `asc finance regions` to list all valid region codes and currencies
- Requires Account Holder, Admin, or Finance role

**Region codes reference:** https://developer.apple.com/help/app-store-connect/reference/financial-report-regions-and-currencies/

### Sandbox Testers

```bash
# List sandbox testers
asc sandbox list

# Filter by email or territory
asc sandbox list --email "tester@example.com"
asc sandbox list --territory "USA"

# Fetch all sandbox testers (all pages)
asc sandbox list --paginate

# Create a sandbox tester
asc sandbox create \
  --email "tester@example.com" \
  --first-name "Test" \
  --last-name "User" \
  --password "Passwordtest1" \
  --confirm-password "Passwordtest1" \
  --secret-question "Question" \
  --secret-answer "Answer" \
  --birth-date "1980-03-01" \
  --territory "USA"

# Get sandbox tester details
asc sandbox get --id "SANDBOX_TESTER_ID"
asc sandbox get --email "tester@example.com"

# Delete a sandbox tester
asc sandbox delete --id "SANDBOX_TESTER_ID" --confirm

# Update a sandbox tester
asc sandbox update --id "SANDBOX_TESTER_ID" --territory "USA"
asc sandbox update --email "tester@example.com" --interrupt-purchases
asc sandbox update --id "SANDBOX_TESTER_ID" --subscription-renewal-rate "MONTHLY_RENEWAL_EVERY_ONE_HOUR"

# Clear purchase history
asc sandbox clear-history --id "SANDBOX_TESTER_ID" --confirm
```

Notes:
- Required create fields: email, first/last name, password + confirm, secret question/answer, birth date, territory
- Password must be 8+ chars with uppercase, lowercase, and a number
- Secret question/answer require 6+ characters
- Territory uses 3-letter App Store territory codes (e.g., `USA`, `JPN`)
- Sandbox list/get use the v2 API; create/delete use v1 endpoints (may be unavailable on some accounts)
- Update/clear-history use the v2 API

### Xcode Cloud

```bash
# List workflows for an app (find workflow IDs)
asc xcode-cloud workflows --app "123456789"

# List build runs for a workflow (find run IDs)
asc xcode-cloud build-runs --workflow-id "WORKFLOW_ID"

# Fetch all workflows/build runs (all pages)
asc xcode-cloud workflows --app "123456789" --paginate
asc xcode-cloud build-runs --workflow-id "WORKFLOW_ID" --paginate

# Trigger a workflow by name (requires --app)
asc xcode-cloud run --app "123456789" --workflow "CI Build" --branch "main"

# Trigger a workflow by ID (no app needed)
asc xcode-cloud run --workflow-id "WORKFLOW_ID" --git-reference-id "REF_ID"

# Trigger and wait for completion
asc xcode-cloud run --app "123456789" --workflow "Deploy" --branch "release/1.0" --wait

# Trigger with custom polling interval and timeout
asc xcode-cloud run --app "123456789" --workflow "CI" --branch "main" --wait --poll-interval 30s --timeout 1h

# Check build run status
asc xcode-cloud status --run-id "BUILD_RUN_ID"

# Check status with table output
asc xcode-cloud status --run-id "BUILD_RUN_ID" --output table

# Wait for an existing build run to complete
asc xcode-cloud status --run-id "BUILD_RUN_ID" --wait
```

Notes:
- Workflows must have manual start conditions enabled to be triggered via API
- Use `--workflow` with `--app` for human-friendly workflow lookup by name
- Use `--workflow-id` and `--git-reference-id` for direct ID-based triggering
- Use `asc xcode-cloud workflows` and `asc xcode-cloud build-runs` to discover IDs
- When using `--wait`, the command polls until the build completes (or times out)
- Exit code is non-zero if the build fails, errors, or is canceled
- Use `ASC_TIMEOUT` env var or `--timeout` flag for long-running builds

### Apps & Builds

```bash
# List apps (useful for finding app IDs)
asc apps

# Sort apps by name or bundle ID
asc apps --sort name
asc apps --sort -bundleId

# Fetch all apps (all pages)
asc apps --paginate

# List builds for an app
asc builds list --app "123456789"

# Sort builds by upload date (newest first)
asc builds list --app "123456789" --sort -uploadedDate

# Fetch all builds (all pages)
asc builds list --app "123456789" --paginate

# Build details
asc builds info --build "BUILD_ID"

# Expire a build (irreversible)
asc builds expire --build "BUILD_ID"

# Prepare a build upload
asc builds upload --app "123456789" --ipa "app.ipa"

Notes:
- Build upload currently prepares upload operations only (upload + commit is not yet automated).

# Add/remove beta groups from a build
asc builds add-groups --build "BUILD_ID" --group "GROUP_ID"
asc builds remove-groups --build "BUILD_ID" --group "GROUP_ID"
```

### Versions

```bash
# List App Store versions
asc versions list --app "123456789"

# Fetch all versions (all pages)
asc versions list --app "123456789" --paginate

# Get version details
asc versions get --version-id "VERSION_ID"

# Create a new version
asc versions create --app "123456789" --version "2.0.0"
asc versions create --app "123456789" --version "2.0.0" --platform IOS --copyright "2026 My Company"

# Update a version
asc versions update --version-id "VERSION_ID" --copyright "2026 My Company"
asc versions update --version-id "VERSION_ID" --release-type SCHEDULED --earliest-release-date "2026-02-01T08:00:00+00:00"

# Delete a version (only PREPARE_FOR_SUBMISSION state)
asc versions delete --version-id "VERSION_ID" --confirm

# Attach a build to a version
asc versions attach-build --version-id "VERSION_ID" --build "BUILD_ID"

# Manage phased release
asc versions phased-release get --version-id "VERSION_ID"
asc versions phased-release create --version-id "VERSION_ID"
asc versions phased-release update --id "PHASED_ID" --state PAUSED
asc versions phased-release delete --id "PHASED_ID" --confirm
```

### Pre-Release Versions

```bash
# List pre-release versions for an app
asc pre-release-versions list --app "123456789"

# Filter by platform or version
asc pre-release-versions list --app "123456789" --platform IOS
asc pre-release-versions list --app "123456789" --version "1.0.0"

# Fetch all pre-release versions (all pages)
asc pre-release-versions list --app "123456789" --paginate

# Get pre-release version details
asc pre-release-versions get --id "PRERELEASE_ID"
```

### Localizations

```bash
# List version localizations
asc localizations list --version "VERSION_ID"

# List app info localizations
asc localizations list --app "APP_ID" --type app-info

# Fetch all localizations (all pages)
asc localizations list --version "VERSION_ID" --paginate

# Download/upload localization files
asc localizations download --version "VERSION_ID" --path "./localizations"
asc localizations upload --version "VERSION_ID" --path "./localizations"
```

### Categories

```bash
# List all App Store categories
asc categories list

# List with table output
asc categories list --output table
```

### Migrate (Fastlane)

Migrate metadata between App Store Connect and fastlane directory structure.

```bash
# Import metadata from fastlane to App Store Connect
asc migrate import --app "APP_ID" --version-id "VERSION_ID" --fastlane-dir ./fastlane

# Preview import without uploading (dry run)
asc migrate import --app "APP_ID" --version-id "VERSION_ID" --fastlane-dir ./fastlane --dry-run

# Export metadata from App Store Connect to fastlane format
asc migrate export --app "APP_ID" --version-id "VERSION_ID" --output-dir ./fastlane

# Validate fastlane metadata without uploading
asc migrate validate --fastlane-dir ./fastlane
asc migrate validate --fastlane-dir ./fastlane --output table
```

**Supported fastlane structure:**
```
fastlane/
├── metadata/
│   ├── en-US/
│   │   ├── name.txt           (App Info - 30 chars max)
│   │   ├── subtitle.txt       (App Info - 30 chars max)
│   │   ├── description.txt    (Version - 4000 chars max)
│   │   ├── keywords.txt       (Version - 100 chars max)
│   │   ├── release_notes.txt  (Version - 4000 chars max)
│   │   ├── promotional_text.txt (Version - 170 chars max)
│   │   ├── support_url.txt    (Version)
│   │   └── marketing_url.txt  (Version)
│   └── de-DE/
│       └── ...
```

Notes:
- `migrate validate` checks character limits before import
- `migrate import` creates new localizations or updates existing ones
- `migrate export` creates the fastlane directory structure from current metadata

### Build Localizations

```bash
# List build release notes localizations
asc build-localizations list --build "BUILD_ID"

# Fetch all build localizations (all pages)
asc build-localizations list --build "BUILD_ID" --paginate

# Create/update/delete release notes
asc build-localizations create --build "BUILD_ID" --locale "en-US" --whats-new "Bug fixes"
asc build-localizations update --id "LOCALIZATION_ID" --whats-new "New features"
asc build-localizations delete --id "LOCALIZATION_ID" --confirm

# Fetch a localization by ID
asc build-localizations get --id "LOCALIZATION_ID"
```

### Submit

```bash
# Submit a build for review
asc submit create --app "123456789" --version "1.0.0" --build "BUILD_ID" --confirm

# Check submission status
asc submit status --id "SUBMISSION_ID"
asc submit status --version-id "VERSION_ID"

# Cancel a submission
asc submit cancel --id "SUBMISSION_ID" --confirm
asc submit cancel --version-id "VERSION_ID" --confirm
```

### Utilities

```bash
# Print version information
asc version
asc --version
```

### Output Formats

| Format | Flag | Use Case |
|--------|------|----------|
| JSON (minified) | default | AI agents, scripting |
| Table | `--output table` | Humans in terminal |
| Markdown | `--output markdown` | Humans, documentation |

Note: When using `--paginate`, the response `links` field is cleared to avoid confusion about additional pages.

### Authentication

```bash
# Check authentication status
asc auth status

# Logout
asc auth logout
```

## Design Philosophy

### Explicit Over Cryptic

```bash
# Good - self-documenting
asc reviews --app "MyApp" --stars 1

# Avoid - cryptic flags (hypothetical, not supported)
# asc reviews -a "MyApp" -s 1
```

### AI-Agent Friendly

All commands output minified JSON by default for easy parsing by AI agents:

```bash
asc feedback --app "123456789" | jq '.data[].attributes.comment'
```

JSON output is minified (one line per response) by default. Use `--output table` or `--output markdown` for human-readable output.

### No Interactive Prompts

Everything is flag-based for automation:

```bash
# Non-interactive (good for CI/CD and AI)
asc feedback --app "123456789"

# No prompts, no waiting
```

## Installation

### Homebrew (macOS)

```bash
# Add tap and install
brew tap rudrankriyam/tap
brew install rudrankriyam/tap/asc
```

### From Source

```bash
git clone https://github.com/rudrankriyam/App-Store-Connect-CLI.git
cd App-Store-Connect-CLI
make build
make install  # Installs to /usr/local/bin
```

## Documentation

- [CLAUDE.md](CLAUDE.md) - Development guidelines for AI assistants
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines

## Security

- Credentials stored in the system keychain when available
- Local config fallback with restricted permissions
- Private key content never stored, only path reference
- Environment variables as fallback

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Author

[Rudrank Riyam](https://github.com/rudrankriyam)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=rudrankriyam/App-Store-Connect-CLI&type=Date)](https://star-history.com/#rudrankriyam/App-Store-Connect-CLI&Date)

---

<p align="center">
  Primarily Built with Cursor and GPT-5.2 Codex Extra High, with MiniMax M2.1 and Claude Code for Implementation
</p>
