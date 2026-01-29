package passtypeids

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// PassTypeIDCertificatesCommand returns the certificates subcommand group.
func PassTypeIDCertificatesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("certificates", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "certificates",
		ShortUsage: "asc pass-type-ids certificates <subcommand> [flags]",
		ShortHelp:  "List pass type ID certificates.",
		LongHelp: `List pass type ID certificates.

Examples:
  asc pass-type-ids certificates list --pass-type-id "PASS_ID"
  asc pass-type-ids certificates get --pass-type-id "PASS_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			PassTypeIDCertificatesListCommand(),
			PassTypeIDCertificatesGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// PassTypeIDCertificatesListCommand returns the certificates list subcommand.
func PassTypeIDCertificatesListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	passTypeID := fs.String("pass-type-id", "", "Pass type ID")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc pass-type-ids certificates list --pass-type-id \"PASS_ID\" [flags]",
		ShortHelp:  "List certificates for a pass type ID.",
		LongHelp: `List certificates for a pass type ID.

Examples:
  asc pass-type-ids certificates list --pass-type-id "PASS_ID"
  asc pass-type-ids certificates list --pass-type-id "PASS_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			passTypeIDValue := strings.TrimSpace(*passTypeID)
			if passTypeIDValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --pass-type-id is required")
				return flag.ErrHelp
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("pass-type-ids certificates list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("pass-type-ids certificates list: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("pass-type-ids certificates list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.PassTypeIDCertificatesOption{
				asc.WithPassTypeIDCertificatesLimit(*limit),
				asc.WithPassTypeIDCertificatesNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithPassTypeIDCertificatesLimit(200))
				firstPage, err := client.GetPassTypeIDCertificates(requestCtx, passTypeIDValue, paginateOpts...)
				if err != nil {
					return fmt.Errorf("pass-type-ids certificates list: failed to fetch: %w", err)
				}

				paginated, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetPassTypeIDCertificates(ctx, passTypeIDValue, asc.WithPassTypeIDCertificatesNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("pass-type-ids certificates list: %w", err)
				}

				return printOutput(paginated, *output, *pretty)
			}

			resp, err := client.GetPassTypeIDCertificates(requestCtx, passTypeIDValue, opts...)
			if err != nil {
				return fmt.Errorf("pass-type-ids certificates list: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// PassTypeIDCertificatesGetCommand returns the certificates get subcommand.
func PassTypeIDCertificatesGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	passTypeID := fs.String("pass-type-id", "", "Pass type ID")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc pass-type-ids certificates get --pass-type-id \"PASS_ID\" [flags]",
		ShortHelp:  "Get certificate relationships for a pass type ID.",
		LongHelp: `Get certificate relationships for a pass type ID.

Examples:
  asc pass-type-ids certificates get --pass-type-id "PASS_ID"
  asc pass-type-ids certificates get --pass-type-id "PASS_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			passTypeIDValue := strings.TrimSpace(*passTypeID)
			if passTypeIDValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --pass-type-id is required")
				return flag.ErrHelp
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("pass-type-ids certificates get: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("pass-type-ids certificates get: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("pass-type-ids certificates get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.LinkagesOption{
				asc.WithLinkagesLimit(*limit),
				asc.WithLinkagesNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithLinkagesLimit(200))
				firstPage, err := client.GetPassTypeIDCertificatesRelationships(requestCtx, passTypeIDValue, paginateOpts...)
				if err != nil {
					return fmt.Errorf("pass-type-ids certificates get: failed to fetch: %w", err)
				}

				paginated, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetPassTypeIDCertificatesRelationships(ctx, passTypeIDValue, asc.WithLinkagesNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("pass-type-ids certificates get: %w", err)
				}

				return printOutput(paginated, *output, *pretty)
			}

			resp, err := client.GetPassTypeIDCertificatesRelationships(requestCtx, passTypeIDValue, opts...)
			if err != nil {
				return fmt.Errorf("pass-type-ids certificates get: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}
