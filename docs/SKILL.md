# ASC CLI Skill for Claude Code

A skill for managing App Store Connect resources using the `asc` command-line tool.

## Quick Reference

```bash
asc --help                     # Discover all commands
asc <command> --help           # Show subcommands and flags
```

## Authentication

Before using the CLI, authentication must be configured:

```bash
# Interactive setup (stores in keychain)
asc auth login

# Or via environment variables
export ASC_KEY_ID="your_key_id"
export ASC_ISSUER_ID="your_issuer_id"
export ASC_PRIVATE_KEY_PATH="/path/to/AuthKey_XXX.p8"
```

API keys are created at https://appstoreconnect.apple.com/access/integrations/api

## Common Workflows

### Discover Apps

```bash
# List all apps
asc apps list

# Get a specific app
asc apps get --app APP_ID

# Set default app for session
export ASC_APP_ID="your_app_id"
```

### Version Management

```bash
# List versions for an app
asc versions list --app APP_ID

# Create a new version
asc versions create --app APP_ID --platform IOS --version "1.2.0"

# Get version details
asc versions get --version-id VERSION_ID

# Delete a version (before submission)
asc versions delete --version-id VERSION_ID --confirm
```

### Localization Management

```bash
# List localizations for a version
asc localizations list --version-id VERSION_ID

# Download all localizations to files
asc localizations download --version-id VERSION_ID --output-dir ./metadata

# Upload localization changes
asc localizations upload --version-id VERSION_ID --locale en-US \
  --description "New description" \
  --keywords "app, utility, tool"

# Create a new localization
asc localizations create --version-id VERSION_ID --locale de-DE

# Delete a localization
asc localizations delete --localization-id LOC_ID --confirm
```

### Fastlane Migration

```bash
# Import metadata from fastlane structure to App Store Connect
asc migrate import --app APP_ID --version-id VERSION_ID --fastlane-dir ./fastlane

# Preview changes without uploading
asc migrate import --app APP_ID --version-id VERSION_ID --fastlane-dir ./fastlane --dry-run

# Export metadata from App Store Connect to fastlane structure
asc migrate export --app APP_ID --version-id VERSION_ID --output-dir ./fastlane
```

Expected fastlane structure:
```
fastlane/
├── metadata/
│   ├── en-US/
│   │   ├── description.txt
│   │   ├── keywords.txt
│   │   ├── release_notes.txt
│   │   ├── promotional_text.txt
│   │   ├── support_url.txt
│   │   ├── marketing_url.txt
│   │   ├── name.txt
│   │   └── subtitle.txt
│   └── de-DE/
│       └── ...
└── screenshots/
    └── en-US/
        └── ...
```

### Screenshots and Previews

```bash
# List screenshot sets for a localization
asc assets screenshots list --localization-id LOC_ID

# Upload a screenshot
asc assets screenshots upload --screenshot-set-id SET_ID --file ./screenshot.png

# Delete a screenshot
asc assets screenshots delete --screenshot-id SCREENSHOT_ID --confirm

# List preview sets
asc assets previews list --localization-id LOC_ID

# Upload a preview video
asc assets previews upload --preview-set-id SET_ID --file ./preview.mp4
```

### TestFlight Management

```bash
# List builds
asc builds list --app APP_ID

# Get latest build
asc builds latest --app APP_ID

# Upload a build (requires App Store Connect API key with upload permissions)
asc builds upload --app APP_ID --file ./App.ipa

# List beta groups
asc beta-groups list --app APP_ID

# Add build to beta group
asc beta-groups add-build --group-id GROUP_ID --build-id BUILD_ID

# List beta testers
asc beta-testers list --app APP_ID

# Invite a beta tester
asc beta-testers create --app APP_ID --email "tester@example.com" --first-name "Test" --last-name "User"
```

### App Categories

```bash
# List available app categories
asc categories list

# Filter by platform
asc categories list --platforms IOS
```

### In-App Purchases

```bash
# List IAPs
asc iap list --app APP_ID

# Get IAP details
asc iap get --iap-id IAP_ID

# Create an IAP
asc iap create --app APP_ID --product-id "com.app.premium" --name "Premium" --type CONSUMABLE
```

### Subscriptions

```bash
# List subscription groups
asc subscriptions groups list --app APP_ID

# List subscriptions in a group
asc subscriptions list --group-id GROUP_ID

# Get subscription details
asc subscriptions get --subscription-id SUB_ID
```

### Pricing

```bash
# List price points
asc pricing points --app APP_ID

# List territories
asc pricing territories
```

### Analytics

```bash
# Request an analytics report
asc analytics request --app APP_ID --report-type APP_USAGE --frequency DAILY

# List report requests
asc analytics requests --app APP_ID

# Download report data
asc analytics download --report-id REPORT_ID --output-dir ./reports
```

### Finance Reports

```bash
# Download sales report
asc finance sales --vendor VENDOR_NUMBER --report-type SALES --date-type DAILY --date "2024-01"

# List available finance regions
asc finance regions --vendor VENDOR_NUMBER
```

### App Submission

```bash
# Check submission readiness
asc submit status --version-id VERSION_ID

# Submit for review
asc submit create --version-id VERSION_ID

# Cancel submission
asc submit cancel --version-id VERSION_ID --confirm

# Manage phased release
asc versions phased-release create --version-id VERSION_ID
asc versions phased-release pause --version-id VERSION_ID
asc versions phased-release resume --version-id VERSION_ID
asc versions phased-release complete --version-id VERSION_ID
```

### Sandbox Testing

```bash
# List sandbox testers
asc sandbox testers list

# Create sandbox tester
asc sandbox testers create --email "sandbox@example.com" --password "SecurePass123" --territory USA

# Clear purchase history
asc sandbox testers clear-history --tester-id TESTER_ID --confirm
```

### Xcode Cloud

```bash
# List CI products
asc xcode-cloud products --app APP_ID

# List workflows
asc xcode-cloud workflows --product-id PRODUCT_ID

# List build runs
asc xcode-cloud runs --workflow-id WORKFLOW_ID

# Start a build
asc xcode-cloud run --workflow-id WORKFLOW_ID
```

## Output Formats

All commands support multiple output formats:

```bash
# JSON (default) - minified for token efficiency
asc apps list

# Pretty JSON - indented for readability
asc apps list --pretty

# Markdown table - good for documentation
asc apps list --output markdown

# Plain table - good for terminal viewing
asc apps list --output table
```

## Pagination

For large result sets:

```bash
# Fetch all pages automatically
asc builds list --app APP_ID --paginate

# Manual pagination with limit
asc builds list --app APP_ID --limit 50

# Continue from cursor
asc builds list --app APP_ID --next "https://api.appstoreconnect.apple.com/..."
```

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `ASC_KEY_ID` | API key ID |
| `ASC_ISSUER_ID` | Issuer ID |
| `ASC_PRIVATE_KEY_PATH` | Path to .p8 key file |
| `ASC_APP_ID` | Default app ID |
| `ASC_VENDOR_NUMBER` | Vendor number for finance reports |
| `ASC_TIMEOUT` | Request timeout (e.g., `90s`, `2m`) |

## Error Handling

The CLI returns structured errors. Common error patterns:

- `NOT_FOUND`: Resource doesn't exist
- `FORBIDDEN`: Insufficient API key permissions
- `CONFLICT`: Resource already exists or in conflicting state
- `INVALID_REQUEST`: Invalid parameters

Always check `--help` for required vs optional flags before constructing commands.

## Best Practices

1. **Always use `--help`** to discover current command structure
2. **Use `--dry-run`** when available to preview changes
3. **Set `ASC_APP_ID`** to avoid repeating `--app` flag
4. **Use JSON output** for programmatic parsing (default)
5. **Use `--confirm`** for destructive operations (delete, cancel)
6. **Use `--paginate`** to fetch complete result sets
