package merchantids

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// MerchantIDsCertificatesCommand returns the merchant ID certificates command with subcommands.
func MerchantIDsCertificatesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("certificates", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "certificates",
		ShortUsage: "asc merchant-ids certificates <subcommand> [flags]",
		ShortHelp:  "List merchant ID certificates.",
		LongHelp: `List merchant ID certificates.

Examples:
  asc merchant-ids certificates list --merchant-id "MERCHANT_ID"
  asc merchant-ids certificates get --merchant-id "MERCHANT_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			MerchantIDsCertificatesListCommand(),
			MerchantIDsCertificatesGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// MerchantIDsCertificatesListCommand returns the certificates list subcommand.
func MerchantIDsCertificatesListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("certificates list", flag.ExitOnError)

	merchantID := fs.String("merchant-id", "", "Merchant ID")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc merchant-ids certificates list --merchant-id \"MERCHANT_ID\" [flags]",
		ShortHelp:  "List certificates for a merchant ID.",
		LongHelp: `List certificates for a merchant ID.

Examples:
  asc merchant-ids certificates list --merchant-id "MERCHANT_ID"
  asc merchant-ids certificates list --merchant-id "MERCHANT_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			merchantIDValue := strings.TrimSpace(*merchantID)
			if merchantIDValue == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --merchant-id is required")
				return flag.ErrHelp
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("merchant-ids certificates list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("merchant-ids certificates list: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids certificates list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.MerchantIDCertificatesOption{
				asc.WithMerchantIDCertificatesLimit(*limit),
				asc.WithMerchantIDCertificatesNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithMerchantIDCertificatesLimit(200))
				firstPage, err := client.GetMerchantIDCertificates(requestCtx, merchantIDValue, paginateOpts...)
				if err != nil {
					return fmt.Errorf("merchant-ids certificates list: failed to fetch: %w", err)
				}

				paginated, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetMerchantIDCertificates(ctx, merchantIDValue, asc.WithMerchantIDCertificatesNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("merchant-ids certificates list: %w", err)
				}

				return printOutput(paginated, *output, *pretty)
			}

			resp, err := client.GetMerchantIDCertificates(requestCtx, merchantIDValue, opts...)
			if err != nil {
				return fmt.Errorf("merchant-ids certificates list: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// MerchantIDsCertificatesGetCommand returns the certificates relationships get subcommand.
func MerchantIDsCertificatesGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("certificates get", flag.ExitOnError)

	merchantID := fs.String("merchant-id", "", "Merchant ID")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc merchant-ids certificates get --merchant-id \"MERCHANT_ID\" [flags]",
		ShortHelp:  "Get certificate relationships for a merchant ID.",
		LongHelp: `Get certificate relationships for a merchant ID.

Examples:
  asc merchant-ids certificates get --merchant-id "MERCHANT_ID"
  asc merchant-ids certificates get --merchant-id "MERCHANT_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			merchantIDValue := strings.TrimSpace(*merchantID)
			if merchantIDValue == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --merchant-id is required")
				return flag.ErrHelp
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("merchant-ids certificates get: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("merchant-ids certificates get: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids certificates get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.LinkagesOption{
				asc.WithLinkagesLimit(*limit),
				asc.WithLinkagesNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithLinkagesLimit(200))
				firstPage, err := client.GetMerchantIDCertificatesRelationships(requestCtx, merchantIDValue, paginateOpts...)
				if err != nil {
					return fmt.Errorf("merchant-ids certificates get: failed to fetch: %w", err)
				}

				paginated, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetMerchantIDCertificatesRelationships(ctx, merchantIDValue, asc.WithLinkagesNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("merchant-ids certificates get: %w", err)
				}

				return printOutput(paginated, *output, *pretty)
			}

			resp, err := client.GetMerchantIDCertificatesRelationships(requestCtx, merchantIDValue, opts...)
			if err != nil {
				return fmt.Errorf("merchant-ids certificates get: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}
