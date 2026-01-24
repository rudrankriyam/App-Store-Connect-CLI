package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
)

// ANSI escape codes for bold text
var (
	bold  = "\033[1m"
	reset = "\033[22m"
)

// Bold returns the string wrapped in ANSI bold codes
func Bold(s string) string {
	return bold + s + reset
}

// DefaultUsageFunc returns a usage string with bold section headers
func DefaultUsageFunc(c *ffcli.Command) string {
	var b strings.Builder

	shortHelp := strings.TrimSpace(c.ShortHelp)
	longHelp := strings.TrimSpace(c.LongHelp)
	if shortHelp == "" && longHelp != "" {
		shortHelp = longHelp
		longHelp = ""
	}

	// DESCRIPTION
	if shortHelp != "" {
		b.WriteString(Bold("DESCRIPTION"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(shortHelp)
		b.WriteString("\n\n")
	}

	// USAGE / ShortUsage
	usage := strings.TrimSpace(c.ShortUsage)
	if usage == "" {
		usage = strings.TrimSpace(c.Name)
	}
	if usage != "" {
		b.WriteString(Bold("USAGE"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(usage)
		b.WriteString("\n\n")
	}

	// LongHelp (additional description)
	if longHelp != "" {
		if shortHelp != "" && strings.HasPrefix(longHelp, shortHelp) {
			longHelp = strings.TrimSpace(strings.TrimPrefix(longHelp, shortHelp))
		}
		if longHelp != "" {
			b.WriteString(longHelp)
			b.WriteString("\n\n")
		}
	}

	// SUBCOMMANDS
	if len(c.Subcommands) > 0 {
		b.WriteString(Bold("SUBCOMMANDS"))
		b.WriteString("\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, sub := range c.Subcommands {
			fmt.Fprintf(tw, "  %-12s %s\n", sub.Name, sub.ShortHelp)
		}
		tw.Flush()
		b.WriteString("\n")
	}

	// FLAGS
	if c.FlagSet != nil {
		hasFlags := false
		c.FlagSet.VisitAll(func(*flag.Flag) {
			hasFlags = true
		})
		if hasFlags {
			b.WriteString(Bold("FLAGS"))
			b.WriteString("\n")
			tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
			c.FlagSet.VisitAll(func(f *flag.Flag) {
				def := f.DefValue
				if def != "" {
					fmt.Fprintf(tw, "  --%-12s %s (default: %s)\n", f.Name, f.Usage, def)
					return
				}
				fmt.Fprintf(tw, "  --%-12s %s\n", f.Name, f.Usage)
			})
			tw.Flush()
			b.WriteString("\n")
		}
	}

	return b.String()
}

func getASCClient() (*asc.Client, error) {
	var actualKeyID, actualIssuerID, actualKeyPath string

	// Priority 1: Keychain credentials (explicit user setup via 'asc auth login')
	cfg, err := auth.GetDefaultCredentials()
	if err == nil && cfg != nil {
		actualKeyID = cfg.KeyID
		actualIssuerID = cfg.IssuerID
		actualKeyPath = cfg.PrivateKeyPath
	}

	// Priority 2: Environment variables (fallback for CI/CD or when keychain unavailable)
	if actualKeyID == "" {
		actualKeyID = os.Getenv("ASC_KEY_ID")
	}
	if actualIssuerID == "" {
		actualIssuerID = os.Getenv("ASC_ISSUER_ID")
	}
	if actualKeyPath == "" {
		actualKeyPath = os.Getenv("ASC_PRIVATE_KEY_PATH")
	}

	if actualKeyID == "" || actualIssuerID == "" || actualKeyPath == "" {
		return nil, fmt.Errorf("missing authentication. Run 'asc auth login'")
	}

	return asc.NewClient(actualKeyID, actualIssuerID, actualKeyPath)
}

func printOutput(data interface{}, format string, pretty bool) error {
	format = strings.ToLower(format)
	switch format {
	case "json":
		if pretty {
			return asc.PrintPrettyJSON(data)
		}
		return asc.PrintJSON(data)
	case "markdown", "md":
		if pretty {
			return fmt.Errorf("--pretty is only valid with JSON output")
		}
		return asc.PrintMarkdown(data)
	case "table":
		if pretty {
			return fmt.Errorf("--pretty is only valid with JSON output")
		}
		return asc.PrintTable(data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func resolveAppID(appID string) string {
	if appID != "" {
		return appID
	}
	return os.Getenv("ASC_APP_ID")
}

func contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, asc.ResolveTimeout())
}

func contextWithUploadTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, asc.ResolveUploadTimeout())
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return cleaned
}

func splitCSVUpper(value string) []string {
	values := splitCSV(value)
	if len(values) == 0 {
		return nil
	}
	upper := make([]string, 0, len(values))
	for _, item := range values {
		upper = append(upper, strings.ToUpper(item))
	}
	return upper
}

func validateNextURL(next string) error {
	next = strings.TrimSpace(next)
	if next == "" {
		return nil
	}
	parsed, err := url.Parse(next)
	if err != nil {
		return fmt.Errorf("--next must be a valid URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "api.appstoreconnect.apple.com" {
		return fmt.Errorf("--next must be an App Store Connect URL")
	}
	return nil
}

func validateSort(value string, allowed ...string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, option := range allowed {
		if value == option {
			return nil
		}
	}
	return fmt.Errorf("--sort must be one of: %s", strings.Join(allowed, ", "))
}
