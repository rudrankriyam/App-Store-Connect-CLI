# Contributing

Thanks for your interest in contributing to ASC (App Store Connect CLI).

## Development Setup

Requirements:
- Go 1.21+

Clone and build:
```bash
git clone https://github.com/rudrankriyam/App-Store-Connect-CLI.git
cd App-Store-Connect-CLI
make build
```

Run tests:
```bash
go test ./...
```

Optional tooling:
```bash
make lint    # uses golangci-lint if installed, else go vet
make format  # gofmt + gofumpt (if gofumpt is installed)
```

## Integration Tests (Opt-in)

Integration tests hit the real App Store Connect API and are skipped by default.
Set credentials in your environment and run:

```bash
export ASC_KEY_ID="YOUR_KEY_ID"
export ASC_ISSUER_ID="YOUR_ISSUER_ID"
export ASC_PRIVATE_KEY_PATH="/path/to/AuthKey.p8"
export ASC_APP_ID="YOUR_APP_ID"

make test-integration
```

## Local API Testing (Optional)

If you have App Store Connect API credentials, you can run real API calls locally:

```bash
export ASC_KEY_ID="YOUR_KEY_ID"
export ASC_ISSUER_ID="YOUR_ISSUER_ID"
export ASC_PRIVATE_KEY_PATH="/path/to/AuthKey.p8"
export ASC_APP_ID="YOUR_APP_ID"

asc feedback --app "$ASC_APP_ID"
asc crashes --app "$ASC_APP_ID"
asc reviews --app "$ASC_APP_ID"
```

Credentials are stored in the system keychain when available, with a local config fallback at
`~/.asc/config.json` (restricted permissions). Do not commit secrets.

## Pull Request Guidelines

- Keep PRs small and focused.
- Add or update tests for new behavior.
- Update `README.md` if behavior or scope changes.
- Avoid committing any credentials or `.p8` files.

## Security

If you find a security issue, please report it responsibly by opening a private issue
or contacting the maintainer directly.
