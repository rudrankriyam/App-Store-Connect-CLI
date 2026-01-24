package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// CategoriesCommand returns the categories command with subcommands.
func CategoriesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("categories", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "categories",
		ShortUsage: "asc categories <subcommand> [flags]",
		ShortHelp:  "Manage App Store categories.",
		LongHelp: `Manage App Store categories.

Examples:
  asc categories list`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			CategoriesListCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// CategoriesListCommand returns the categories list subcommand.
func CategoriesListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("categories list", flag.ExitOnError)

	limit := fs.Int("limit", 200, "Maximum results to fetch (1-200)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc categories list [flags]",
		ShortHelp:  "List available App Store categories.",
		LongHelp: `List available App Store categories.

Category IDs can be used when updating app information to set primary
and secondary categories.

Examples:
  asc categories list
  asc categories list --output table`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit < 1 || *limit > 200 {
				return fmt.Errorf("categories list: --limit must be between 1 and 200")
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("categories list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			categories, err := client.GetAppCategories(requestCtx, asc.WithAppCategoriesLimit(*limit))
			if err != nil {
				return fmt.Errorf("categories list: %w", err)
			}

			return printOutput(categories, *output, *pretty)
		},
	}
}
