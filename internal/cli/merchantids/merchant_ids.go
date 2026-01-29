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

// MerchantIDsCommand returns the merchant IDs command with subcommands.
func MerchantIDsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("merchant-ids", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "merchant-ids",
		ShortUsage: "asc merchant-ids <subcommand> [flags]",
		ShortHelp:  "Manage merchant IDs and certificates.",
		LongHelp: `Manage merchant IDs and certificates.

Examples:
  asc merchant-ids list
  asc merchant-ids get --merchant-id "MERCHANT_ID"
  asc merchant-ids create --identifier "merchant.com.example" --name "Example"
  asc merchant-ids update --merchant-id "MERCHANT_ID" --name "New Name"
  asc merchant-ids delete --merchant-id "MERCHANT_ID" --confirm
  asc merchant-ids certificates list --merchant-id "MERCHANT_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			MerchantIDsListCommand(),
			MerchantIDsGetCommand(),
			MerchantIDsCreateCommand(),
			MerchantIDsUpdateCommand(),
			MerchantIDsDeleteCommand(),
			MerchantIDsCertificatesCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// MerchantIDsListCommand returns the merchant IDs list subcommand.
func MerchantIDsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	identifier := fs.String("identifier", "", "Filter by merchant ID identifier(s), comma-separated")
	name := fs.String("name", "", "Filter by merchant ID name(s), comma-separated")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc merchant-ids list [flags]",
		ShortHelp:  "List merchant IDs.",
		LongHelp: `List merchant IDs.

Examples:
  asc merchant-ids list
  asc merchant-ids list --identifier "merchant.com.example"
  asc merchant-ids list --name "Example"
  asc merchant-ids list --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("merchant-ids list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("merchant-ids list: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.MerchantIDsOption{
				asc.WithMerchantIDsLimit(*limit),
				asc.WithMerchantIDsNextURL(*next),
			}
			if strings.TrimSpace(*identifier) != "" {
				opts = append(opts, asc.WithMerchantIDsFilterIdentifier(*identifier))
			}
			if strings.TrimSpace(*name) != "" {
				opts = append(opts, asc.WithMerchantIDsFilterName(*name))
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithMerchantIDsLimit(200))
				firstPage, err := client.GetMerchantIDs(requestCtx, paginateOpts...)
				if err != nil {
					return fmt.Errorf("merchant-ids list: failed to fetch: %w", err)
				}

				paginated, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetMerchantIDs(ctx, asc.WithMerchantIDsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("merchant-ids list: %w", err)
				}

				return printOutput(paginated, *output, *pretty)
			}

			resp, err := client.GetMerchantIDs(requestCtx, opts...)
			if err != nil {
				return fmt.Errorf("merchant-ids list: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// MerchantIDsGetCommand returns the merchant IDs get subcommand.
func MerchantIDsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	merchantID := fs.String("merchant-id", "", "Merchant ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc merchant-ids get --merchant-id \"MERCHANT_ID\"",
		ShortHelp:  "Get a merchant ID by ID.",
		LongHelp: `Get a merchant ID by ID.

Examples:
  asc merchant-ids get --merchant-id "MERCHANT_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			merchantIDValue := strings.TrimSpace(*merchantID)
			if merchantIDValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --merchant-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetMerchantID(requestCtx, merchantIDValue)
			if err != nil {
				return fmt.Errorf("merchant-ids get: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// MerchantIDsCreateCommand returns the merchant IDs create subcommand.
func MerchantIDsCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ExitOnError)

	identifier := fs.String("identifier", "", "Merchant ID identifier (e.g., merchant.com.example)")
	name := fs.String("name", "", "Merchant ID name")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc merchant-ids create --identifier \"merchant.com.example\" --name \"Example\"",
		ShortHelp:  "Create a merchant ID.",
		LongHelp: `Create a merchant ID.

Examples:
  asc merchant-ids create --identifier "merchant.com.example" --name "Example"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			identifierValue := strings.TrimSpace(*identifier)
			if identifierValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --identifier is required")
				return flag.ErrHelp
			}
			nameValue := strings.TrimSpace(*name)
			if nameValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --name is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids create: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			attrs := asc.MerchantIDCreateAttributes{
				Name:       nameValue,
				Identifier: identifierValue,
			}
			resp, err := client.CreateMerchantID(requestCtx, attrs)
			if err != nil {
				return fmt.Errorf("merchant-ids create: failed to create: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// MerchantIDsUpdateCommand returns the merchant IDs update subcommand.
func MerchantIDsUpdateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ExitOnError)

	merchantID := fs.String("merchant-id", "", "Merchant ID")
	name := fs.String("name", "", "Merchant ID name")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "asc merchant-ids update --merchant-id \"MERCHANT_ID\" --name \"New Name\"",
		ShortHelp:  "Update a merchant ID.",
		LongHelp: `Update a merchant ID.

Examples:
  asc merchant-ids update --merchant-id "MERCHANT_ID" --name "New Name"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			merchantIDValue := strings.TrimSpace(*merchantID)
			if merchantIDValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --merchant-id is required")
				return flag.ErrHelp
			}
			nameValue := strings.TrimSpace(*name)
			if nameValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --name is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids update: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			attrs := asc.MerchantIDUpdateAttributes{Name: nameValue}
			resp, err := client.UpdateMerchantID(requestCtx, merchantIDValue, attrs)
			if err != nil {
				return fmt.Errorf("merchant-ids update: failed to update: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// MerchantIDsDeleteCommand returns the merchant IDs delete subcommand.
func MerchantIDsDeleteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	merchantID := fs.String("merchant-id", "", "Merchant ID")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "asc merchant-ids delete --merchant-id \"MERCHANT_ID\" --confirm",
		ShortHelp:  "Delete a merchant ID.",
		LongHelp: `Delete a merchant ID.

Examples:
  asc merchant-ids delete --merchant-id "MERCHANT_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			merchantIDValue := strings.TrimSpace(*merchantID)
			if merchantIDValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --merchant-id is required")
				return flag.ErrHelp
			}
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("merchant-ids delete: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.DeleteMerchantID(requestCtx, merchantIDValue); err != nil {
				return fmt.Errorf("merchant-ids delete: failed to delete: %w", err)
			}

			result := &asc.MerchantIDDeleteResult{
				ID:      merchantIDValue,
				Deleted: true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}
