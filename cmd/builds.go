package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// BuildsAddGroupsCommand returns the builds add-groups subcommand.
func BuildsAddGroupsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("add-groups", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID")
	groups := fs.String("group", "", "Comma-separated beta group IDs")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "add-groups",
		ShortUsage: "asc builds add-groups --build BUILD_ID --group GROUP_ID[,GROUP_ID...]",
		ShortHelp:  "Add beta groups to a build for TestFlight distribution.",
		LongHelp: `Add beta groups to a build for TestFlight distribution.

Examples:
  asc builds add-groups --build "BUILD_ID" --group "GROUP_ID"
  asc builds add-groups --build "BUILD_ID" --group "GROUP1,GROUP2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedBuildID := strings.TrimSpace(*buildID)
			if trimmedBuildID == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}

			groupIDs := splitCSV(*groups)
			if len(groupIDs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds add-groups: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.AddBetaGroupsToBuild(requestCtx, trimmedBuildID, groupIDs); err != nil {
				return fmt.Errorf("builds add-groups: failed to add groups: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Successfully added %d group(s) to build %s\n", len(groupIDs), trimmedBuildID)
			result := &asc.BuildBetaGroupsUpdateResult{
				BuildID:  trimmedBuildID,
				GroupIDs: groupIDs,
				Action:   "added",
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

// BuildsRemoveGroupsCommand returns the builds remove-groups subcommand.
func BuildsRemoveGroupsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("remove-groups", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID")
	groups := fs.String("group", "", "Comma-separated beta group IDs")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "remove-groups",
		ShortUsage: "asc builds remove-groups --build BUILD_ID --group GROUP_ID[,GROUP_ID...]",
		ShortHelp:  "Remove beta groups from a build.",
		LongHelp: `Remove beta groups from a build.

Examples:
  asc builds remove-groups --build "BUILD_ID" --group "GROUP_ID"
  asc builds remove-groups --build "BUILD_ID" --group "GROUP1,GROUP2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedBuildID := strings.TrimSpace(*buildID)
			if trimmedBuildID == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}

			groupIDs := splitCSV(*groups)
			if len(groupIDs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds remove-groups: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.RemoveBetaGroupsFromBuild(requestCtx, trimmedBuildID, groupIDs); err != nil {
				return fmt.Errorf("builds remove-groups: failed to remove groups: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Successfully removed %d group(s) from build %s\n", len(groupIDs), trimmedBuildID)
			result := &asc.BuildBetaGroupsUpdateResult{
				BuildID:  trimmedBuildID,
				GroupIDs: groupIDs,
				Action:   "removed",
			}

			return printOutput(result, *output, *pretty)
		},
	}
}
