package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

var errBetaTesterNotFound = errors.New("beta tester not found")

// BetaGroupsCommand returns the beta groups command with subcommands.
func BetaGroupsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("beta-groups", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "beta-groups",
		ShortUsage: "asc beta-groups <subcommand> [flags]",
		ShortHelp:  "Manage TestFlight beta groups.",
		LongHelp: `Manage TestFlight beta groups.

Examples:
  asc beta-groups list --app "APP_ID"
  asc beta-groups create --app "APP_ID" --name "Beta Testers"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaGroupsListCommand(),
			BetaGroupsCreateCommand(),
			BetaGroupsGetCommand(),
			BetaGroupsUpdateCommand(),
			BetaGroupsAddTestersCommand(),
			BetaGroupsRemoveTestersCommand(),
			BetaGroupsDeleteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BetaGroupsListCommand returns the beta groups list subcommand.
func BetaGroupsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc beta-groups list [flags]",
		ShortHelp:  "List TestFlight beta groups for an app.",
		LongHelp: `List TestFlight beta groups for an app.

Examples:
  asc beta-groups list --app "APP_ID"
  asc beta-groups list --app "APP_ID" --limit 10
  asc beta-groups list --app "APP_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("beta-groups list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("beta-groups list: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.BetaGroupsOption{
				asc.WithBetaGroupsLimit(*limit),
				asc.WithBetaGroupsNextURL(*next),
			}

			if *paginate {
				// Fetch first page with limit set for consistent pagination
				paginateOpts := append(opts, asc.WithBetaGroupsLimit(200))
				firstPage, err := client.GetBetaGroups(requestCtx, resolvedAppID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("beta-groups list: failed to fetch: %w", err)
				}

				// Fetch all remaining pages
				groups, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetBetaGroups(ctx, resolvedAppID, asc.WithBetaGroupsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("beta-groups list: %w", err)
				}

				return printOutput(groups, *output, *pretty)
			}

			groups, err := client.GetBetaGroups(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("beta-groups list: failed to fetch: %w", err)
			}

			return printOutput(groups, *output, *pretty)
		},
	}
}

// BetaGroupsCreateCommand returns the beta groups create subcommand.
func BetaGroupsCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	name := fs.String("name", "", "Beta group name")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc beta-groups create [flags]",
		ShortHelp:  "Create a TestFlight beta group.",
		LongHelp: `Create a TestFlight beta group.

Examples:
  asc beta-groups create --app "APP_ID" --name "Beta Testers"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*name) == "" {
				fmt.Fprintln(os.Stderr, "Error: --name is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups create: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			group, err := client.CreateBetaGroup(requestCtx, resolvedAppID, strings.TrimSpace(*name))
			if err != nil {
				return fmt.Errorf("beta-groups create: failed to create: %w", err)
			}

			return printOutput(group, *output, *pretty)
		},
	}
}

// BetaGroupsGetCommand returns the beta groups get subcommand.
func BetaGroupsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	id := fs.String("id", "", "Beta group ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc beta-groups get [flags]",
		ShortHelp:  "Get a TestFlight beta group by ID.",
		LongHelp: `Get a TestFlight beta group by ID.

Examples:
  asc beta-groups get --id "GROUP_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*id) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			group, err := client.GetBetaGroup(requestCtx, strings.TrimSpace(*id))
			if err != nil {
				return fmt.Errorf("beta-groups get: failed to fetch: %w", err)
			}

			return printOutput(group, *output, *pretty)
		},
	}
}

// BetaGroupsUpdateCommand returns the beta groups update subcommand.
func BetaGroupsUpdateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("update", flag.ExitOnError)

	id := fs.String("id", "", "Beta group ID")
	name := fs.String("name", "", "Beta group name")
	publicLinkEnabled := fs.Bool("public-link-enabled", false, "Enable public link")
	publicLinkLimitEnabled := fs.Bool("public-link-limit-enabled", false, "Enable public link limit")
	publicLinkLimit := fs.Int("public-link-limit", 0, "Public link limit (1-10000)")
	feedbackEnabled := fs.Bool("feedback-enabled", false, "Enable feedback")
	internal := fs.Bool("internal", false, "Set as internal group")
	allBuilds := fs.Bool("all-builds", false, "Grant access to all builds")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "asc beta-groups update [flags]",
		ShortHelp:  "Update a TestFlight beta group.",
		LongHelp: `Update a TestFlight beta group.

Examples:
  asc beta-groups update --id "GROUP_ID" --name "New Name"
  asc beta-groups update --id "GROUP_ID" --public-link-enabled --public-link-limit 100
  asc beta-groups update --id "GROUP_ID" --feedback-enabled`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*id) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			// Use a map to track which flags were explicitly set
			flagSet := make(map[string]bool)
			fs.Visit(func(f *flag.Flag) {
				flagSet[f.Name] = true
			})

			// Validate update flags before creating client
			attrs := asc.BetaGroupUpdateAttributes{}
			hasUpdates := false

			if strings.TrimSpace(*name) != "" {
				attrs.Name = strings.TrimSpace(*name)
				hasUpdates = true
			}

			if flagSet["public-link-enabled"] {
				publicLinkEnabledVal := *publicLinkEnabled
				attrs.PublicLinkEnabled = &publicLinkEnabledVal
				hasUpdates = true
			}

			if flagSet["public-link-limit-enabled"] {
				publicLinkLimitEnabledVal := *publicLinkLimitEnabled
				attrs.PublicLinkLimitEnabled = &publicLinkLimitEnabledVal
				hasUpdates = true
			}

			if flagSet["public-link-limit"] {
				if *publicLinkLimit < 1 || *publicLinkLimit > 10000 {
					fmt.Fprintf(os.Stderr, "Error: --public-link-limit must be between 1 and 10000\n\n")
					return flag.ErrHelp
				}
				attrs.PublicLinkLimit = *publicLinkLimit
				hasUpdates = true
			}

			if flagSet["feedback-enabled"] {
				feedbackEnabledVal := *feedbackEnabled
				attrs.FeedbackEnabled = &feedbackEnabledVal
				hasUpdates = true
			}

			if flagSet["internal"] {
				internalVal := *internal
				attrs.IsInternalGroup = &internalVal
				hasUpdates = true
			}

			if flagSet["all-builds"] {
				allBuildsVal := *allBuilds
				attrs.HasAccessToAllBuilds = &allBuildsVal
				hasUpdates = true
			}

			if !hasUpdates {
				fmt.Fprintln(os.Stderr, "Error: at least one update flag is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups update: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			req := asc.BetaGroupUpdateRequest{
				Data: asc.BetaGroupUpdateData{
					Type:       asc.ResourceTypeBetaGroups,
					ID:         strings.TrimSpace(*id),
					Attributes: &attrs,
				},
			}

			group, err := client.UpdateBetaGroup(requestCtx, strings.TrimSpace(*id), req)
			if err != nil {
				return fmt.Errorf("beta-groups update: failed to update: %w", err)
			}

			return printOutput(group, *output, *pretty)
		},
	}
}

// BetaGroupsDeleteCommand returns the beta groups delete subcommand.
func BetaGroupsDeleteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	id := fs.String("id", "", "Beta group ID")
	confirm := fs.Bool("confirm", false, "Confirm deletion")

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "asc beta-groups delete [flags]",
		ShortHelp:  "Delete a TestFlight beta group.",
		LongHelp: `Delete a TestFlight beta group.

Examples:
  asc beta-groups delete --id "GROUP_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*id) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to delete a beta group")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups delete: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.DeleteBetaGroup(requestCtx, strings.TrimSpace(*id)); err != nil {
				return fmt.Errorf("beta-groups delete: failed to delete: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Successfully deleted beta group %s\n", strings.TrimSpace(*id))
			return nil
		},
	}
}

// BetaGroupsAddTestersCommand returns the beta groups add-testers subcommand.
func BetaGroupsAddTestersCommand() *ffcli.Command {
	fs := flag.NewFlagSet("add-testers", flag.ExitOnError)

	group := fs.String("group", "", "Beta group ID")
	tester := fs.String("tester", "", "Beta tester ID(s), comma-separated")

	return &ffcli.Command{
		Name:       "add-testers",
		ShortUsage: "asc beta-groups add-testers --group \"GROUP_ID\" --tester \"TESTER_ID[,TESTER_ID...]\"",
		ShortHelp:  "Add beta testers to a beta group.",
		LongHelp: `Add beta testers to a beta group.

Examples:
  asc beta-groups add-testers --group "GROUP_ID" --tester "TESTER_ID"
  asc beta-groups add-testers --group "GROUP_ID" --tester "TESTER_ID1,TESTER_ID2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			groupID := strings.TrimSpace(*group)
			if groupID == "" {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			testerIDs := splitCSV(*tester)
			if len(testerIDs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --tester is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups add-testers: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.AddBetaTestersToGroup(requestCtx, groupID, testerIDs); err != nil {
				return fmt.Errorf("beta-groups add-testers: failed to add testers: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Successfully added %d tester(s) to group %s\n", len(testerIDs), groupID)
			return nil
		},
	}
}

// BetaGroupsRemoveTestersCommand returns the beta groups remove-testers subcommand.
func BetaGroupsRemoveTestersCommand() *ffcli.Command {
	fs := flag.NewFlagSet("remove-testers", flag.ExitOnError)

	group := fs.String("group", "", "Beta group ID")
	tester := fs.String("tester", "", "Beta tester ID(s), comma-separated")

	return &ffcli.Command{
		Name:       "remove-testers",
		ShortUsage: "asc beta-groups remove-testers --group \"GROUP_ID\" --tester \"TESTER_ID[,TESTER_ID...]\"",
		ShortHelp:  "Remove beta testers from a beta group.",
		LongHelp: `Remove beta testers from a beta group.

Examples:
  asc beta-groups remove-testers --group "GROUP_ID" --tester "TESTER_ID"
  asc beta-groups remove-testers --group "GROUP_ID" --tester "TESTER_ID1,TESTER_ID2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			groupID := strings.TrimSpace(*group)
			if groupID == "" {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			testerIDs := splitCSV(*tester)
			if len(testerIDs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --tester is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-groups remove-testers: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.RemoveBetaTestersFromGroup(requestCtx, groupID, testerIDs); err != nil {
				return fmt.Errorf("beta-groups remove-testers: failed to remove testers: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Successfully removed %d tester(s) from group %s\n", len(testerIDs), groupID)
			return nil
		},
	}
}

// BetaTestersCommand returns the beta testers command with subcommands.
func BetaTestersCommand() *ffcli.Command {
	fs := flag.NewFlagSet("beta-testers", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "beta-testers",
		ShortUsage: "asc beta-testers <subcommand> [flags]",
		ShortHelp:  "Manage TestFlight beta testers.",
		LongHelp: `Manage TestFlight beta testers.

Examples:
  asc beta-testers list --app "APP_ID"
  asc beta-testers get --id "TESTER_ID"
  asc beta-testers add --app "APP_ID" --email "tester@example.com" --group "Beta"
  asc beta-testers remove --app "APP_ID" --email "tester@example.com"
  asc beta-testers add-groups --id "TESTER_ID" --group "GROUP_ID"
  asc beta-testers remove-groups --id "TESTER_ID" --group "GROUP_ID"
  asc beta-testers invite --app "APP_ID" --email "tester@example.com"
  asc beta-testers invite --app "APP_ID" --email "tester@example.com" --group "Beta"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaTestersListCommand(),
			BetaTestersGetCommand(),
			BetaTestersAddCommand(),
			BetaTestersRemoveCommand(),
			BetaTestersAddGroupsCommand(),
			BetaTestersRemoveGroupsCommand(),
			BetaTestersInviteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BetaTestersListCommand returns the beta testers list subcommand.
func BetaTestersListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	buildID := fs.String("build", "", "Build ID to filter")
	group := fs.String("group", "", "Beta group name or ID to filter")
	email := fs.String("email", "", "Filter by tester email")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc beta-testers list [flags]",
		ShortHelp:  "List TestFlight beta testers for an app.",
		LongHelp: `List TestFlight beta testers for an app.

Examples:
  asc beta-testers list --app "APP_ID"
  asc beta-testers list --app "APP_ID" --build "BUILD_ID"
  asc beta-testers list --app "APP_ID" --group "Beta"
  asc beta-testers list --app "APP_ID" --limit 25
  asc beta-testers list --app "APP_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("beta-testers list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("beta-testers list: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.BetaTestersOption{
				asc.WithBetaTestersLimit(*limit),
				asc.WithBetaTestersNextURL(*next),
			}

			if strings.TrimSpace(*buildID) != "" {
				opts = append(opts, asc.WithBetaTestersBuildID(strings.TrimSpace(*buildID)))
			}

			if strings.TrimSpace(*email) != "" {
				opts = append(opts, asc.WithBetaTestersEmail(*email))
			}

			if strings.TrimSpace(*group) != "" && strings.TrimSpace(*next) == "" {
				groupID, err := resolveBetaGroupID(requestCtx, client, resolvedAppID, *group)
				if err != nil {
					return fmt.Errorf("beta-testers list: %w", err)
				}
				opts = append(opts, asc.WithBetaTestersGroupIDs([]string{groupID}))
			}

			if *paginate {
				// Fetch first page with limit set for consistent pagination
				paginateOpts := append(opts, asc.WithBetaTestersLimit(200))
				firstPage, err := client.GetBetaTesters(requestCtx, resolvedAppID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("beta-testers list: failed to fetch: %w", err)
				}

				// Fetch all remaining pages
				testers, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetBetaTesters(ctx, resolvedAppID, asc.WithBetaTestersNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("beta-testers list: %w", err)
				}

				return printOutput(testers, *output, *pretty)
			}

			testers, err := client.GetBetaTesters(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("beta-testers list: failed to fetch: %w", err)
			}

			return printOutput(testers, *output, *pretty)
		},
	}
}

// BetaTestersGetCommand returns the beta testers get subcommand.
func BetaTestersGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	id := fs.String("id", "", "Beta tester ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc beta-testers get [flags]",
		ShortHelp:  "Get a TestFlight beta tester by ID.",
		LongHelp: `Get a TestFlight beta tester by ID.

Examples:
  asc beta-testers get --id "TESTER_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			tester, err := client.GetBetaTester(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("beta-testers get: failed to fetch: %w", err)
			}

			return printOutput(tester, *output, *pretty)
		},
	}
}

// BetaTestersAddCommand returns the beta testers add subcommand.
func BetaTestersAddCommand() *ffcli.Command {
	fs := flag.NewFlagSet("add", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	email := fs.String("email", "", "Tester email address")
	firstName := fs.String("first-name", "", "Tester first name")
	lastName := fs.String("last-name", "", "Tester last name")
	group := fs.String("group", "", "Beta group name or ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "add",
		ShortUsage: "asc beta-testers add [flags]",
		ShortHelp:  "Add a TestFlight beta tester.",
		LongHelp: `Add a TestFlight beta tester.

Examples:
  asc beta-testers add --app "APP_ID" --email "tester@example.com" --group "Beta"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --email is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*group) == "" {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers add: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			groupID, err := resolveBetaGroupID(requestCtx, client, resolvedAppID, *group)
			if err != nil {
				return fmt.Errorf("beta-testers add: %w", err)
			}

			tester, err := client.CreateBetaTester(requestCtx, *email, *firstName, *lastName, []string{groupID})
			if err != nil {
				return fmt.Errorf("beta-testers add: failed to create: %w", err)
			}

			return printOutput(tester, *output, *pretty)
		},
	}
}

// BetaTestersRemoveCommand returns the beta testers remove subcommand.
func BetaTestersRemoveCommand() *ffcli.Command {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	email := fs.String("email", "", "Tester email address")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "remove",
		ShortUsage: "asc beta-testers remove [flags]",
		ShortHelp:  "Remove a TestFlight beta tester.",
		LongHelp: `Remove a TestFlight beta tester.

Examples:
  asc beta-testers remove --app "APP_ID" --email "tester@example.com"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --email is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers remove: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			testerID, err := findBetaTesterIDByEmail(requestCtx, client, resolvedAppID, *email)
			if err != nil {
				if errors.Is(err, errBetaTesterNotFound) {
					return fmt.Errorf("beta-testers remove: no tester found for %q", strings.TrimSpace(*email))
				}
				return fmt.Errorf("beta-testers remove: %w", err)
			}

			if err := client.DeleteBetaTester(requestCtx, testerID); err != nil {
				return fmt.Errorf("beta-testers remove: failed to remove: %w", err)
			}

			result := &asc.BetaTesterDeleteResult{
				ID:      testerID,
				Email:   strings.TrimSpace(*email),
				Deleted: true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

// BetaTestersAddGroupsCommand returns the beta testers add-groups subcommand.
func BetaTestersAddGroupsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("add-groups", flag.ExitOnError)

	id := fs.String("id", "", "Beta tester ID")
	groups := fs.String("group", "", "Comma-separated beta group IDs")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "add-groups",
		ShortUsage: "asc beta-testers add-groups --id TESTER_ID --group GROUP_ID[,GROUP_ID...]",
		ShortHelp:  "Add a beta tester to beta groups.",
		LongHelp: `Add a beta tester to beta groups.

Examples:
  asc beta-testers add-groups --id "TESTER_ID" --group "GROUP_ID"
  asc beta-testers add-groups --id "TESTER_ID" --group "GROUP_ID_1,GROUP_ID_2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			testerID := strings.TrimSpace(*id)
			if testerID == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			groupIDs := splitCSV(*groups)
			if len(groupIDs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers add-groups: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.AddBetaTesterToGroups(requestCtx, testerID, groupIDs); err != nil {
				return fmt.Errorf("beta-testers add-groups: failed to add groups: %w", err)
			}

			result := &asc.BetaTesterGroupsUpdateResult{
				TesterID: testerID,
				GroupIDs: groupIDs,
				Action:   "added",
			}

			if err := printOutput(result, *output, *pretty); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Successfully added tester %s to %d group(s)\n", testerID, len(groupIDs))
			return nil
		},
	}
}

// BetaTestersRemoveGroupsCommand returns the beta testers remove-groups subcommand.
func BetaTestersRemoveGroupsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("remove-groups", flag.ExitOnError)

	id := fs.String("id", "", "Beta tester ID")
	groups := fs.String("group", "", "Comma-separated beta group IDs")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "remove-groups",
		ShortUsage: "asc beta-testers remove-groups --id TESTER_ID --group GROUP_ID[,GROUP_ID...]",
		ShortHelp:  "Remove a beta tester from beta groups.",
		LongHelp: `Remove a beta tester from beta groups.

Examples:
  asc beta-testers remove-groups --id "TESTER_ID" --group "GROUP_ID"
  asc beta-testers remove-groups --id "TESTER_ID" --group "GROUP_ID_1,GROUP_ID_2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			testerID := strings.TrimSpace(*id)
			if testerID == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			groupIDs := splitCSV(*groups)
			if len(groupIDs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --group is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers remove-groups: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.RemoveBetaTesterFromGroups(requestCtx, testerID, groupIDs); err != nil {
				return fmt.Errorf("beta-testers remove-groups: failed to remove groups: %w", err)
			}

			result := &asc.BetaTesterGroupsUpdateResult{
				TesterID: testerID,
				GroupIDs: groupIDs,
				Action:   "removed",
			}

			if err := printOutput(result, *output, *pretty); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Successfully removed tester %s from %d group(s)\n", testerID, len(groupIDs))
			return nil
		},
	}
}

// BetaTestersInviteCommand returns the beta testers invite subcommand.
func BetaTestersInviteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("invite", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	email := fs.String("email", "", "Tester email address")
	group := fs.String("group", "", "Beta group name or ID (optional, creates tester if missing)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "invite",
		ShortUsage: "asc beta-testers invite [flags]",
		ShortHelp:  "Invite a TestFlight beta tester.",
		LongHelp: `Invite a TestFlight beta tester.

Examples:
  asc beta-testers invite --app "APP_ID" --email "tester@example.com"
  asc beta-testers invite --app "APP_ID" --email "tester@example.com" --group "Beta"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --email is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("beta-testers invite: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			emailValue := strings.TrimSpace(*email)
			groupValue := strings.TrimSpace(*group)
			testerID, err := findBetaTesterIDByEmail(requestCtx, client, resolvedAppID, emailValue)
			if err != nil {
				if errors.Is(err, errBetaTesterNotFound) {
					if groupValue == "" {
						return fmt.Errorf("beta-testers invite: no tester found for %q (use beta-testers add --group ... or pass --group here)", emailValue)
					}

					groupID, resolveErr := resolveBetaGroupID(requestCtx, client, resolvedAppID, groupValue)
					if resolveErr != nil {
						return fmt.Errorf("beta-testers invite: %w", resolveErr)
					}

					created, createErr := client.CreateBetaTester(requestCtx, emailValue, "", "", []string{groupID})
					if createErr != nil {
						return fmt.Errorf("beta-testers invite: failed to create tester: %w", createErr)
					}
					testerID = created.Data.ID
				} else {
					return fmt.Errorf("beta-testers invite: %w", err)
				}
			}

			invitation, err := client.CreateBetaTesterInvitation(requestCtx, resolvedAppID, testerID)
			if err != nil {
				return fmt.Errorf("beta-testers invite: failed to create invitation: %w", err)
			}

			result := &asc.BetaTesterInvitationResult{
				InvitationID: invitation.Data.ID,
				TesterID:     testerID,
				AppID:        resolvedAppID,
				Email:        emailValue,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

func resolveBetaGroupID(ctx context.Context, client *asc.Client, appID, group string) (string, error) {
	group = strings.TrimSpace(group)
	if group == "" {
		return "", fmt.Errorf("beta group name is required")
	}

	groups, err := client.GetBetaGroups(ctx, appID, asc.WithBetaGroupsLimit(200))
	if err != nil {
		return "", err
	}

	for _, item := range groups.Data {
		if item.ID == group {
			return item.ID, nil
		}
	}

	matches := make([]string, 0, 1)
	for _, item := range groups.Data {
		if strings.EqualFold(strings.TrimSpace(item.Attributes.Name), group) {
			matches = append(matches, item.ID)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("beta group %q not found", group)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple beta groups named %q; use group ID", group)
	}
}

func findBetaTesterIDByEmail(ctx context.Context, client *asc.Client, appID, email string) (string, error) {
	testers, err := client.GetBetaTesters(ctx, appID, asc.WithBetaTestersEmail(email))
	if err != nil {
		return "", err
	}

	if len(testers.Data) == 0 {
		return "", errBetaTesterNotFound
	}
	if len(testers.Data) > 1 {
		return "", fmt.Errorf("multiple beta testers found for %q", strings.TrimSpace(email))
	}

	return testers.Data[0].ID, nil
}

