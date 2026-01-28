package cmd

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// MarketplaceCommand returns the marketplace command with subcommands.
func MarketplaceCommand() *ffcli.Command {
	fs := flag.NewFlagSet("marketplace", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "marketplace",
		ShortUsage: "asc marketplace <subcommand> [flags]",
		ShortHelp:  "Manage marketplace resources.",
		LongHelp: `Manage marketplace resources.

Examples:
  asc marketplace search-details get --app "APP_ID"
  asc marketplace webhooks list`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			MarketplaceSearchDetailsCommand(),
			MarketplaceWebhooksCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}
