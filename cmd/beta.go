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

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc beta-groups list [flags]",
		ShortHelp:  "List TestFlight beta groups for an app.",
		LongHelp: `List TestFlight beta groups for an app.

Examples:
  asc beta-groups list --app "APP_ID"
  asc beta-groups list --app "APP_ID" --limit 10`,
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
  asc beta-testers add --app "APP_ID" --email "tester@example.com" --group "Beta"
  asc beta-testers remove --app "APP_ID" --email "tester@example.com"
  asc beta-testers invite --app "APP_ID" --email "tester@example.com"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaTestersListCommand(),
			BetaTestersAddCommand(),
			BetaTestersRemoveCommand(),
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
	group := fs.String("group", "", "Beta group name or ID to filter")
	email := fs.String("email", "", "Filter by tester email")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc beta-testers list [flags]",
		ShortHelp:  "List TestFlight beta testers for an app.",
		LongHelp: `List TestFlight beta testers for an app.

Examples:
  asc beta-testers list --app "APP_ID"
  asc beta-testers list --app "APP_ID" --group "Beta"
  asc beta-testers list --app "APP_ID" --limit 25`,
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

			testers, err := client.GetBetaTesters(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("beta-testers list: failed to fetch: %w", err)
			}

			return printOutput(testers, *output, *pretty)
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

			result := asc.BetaTesterDeleteResult{
				ID:      testerID,
				Email:   strings.TrimSpace(*email),
				Deleted: true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

// BetaTestersInviteCommand returns the beta testers invite subcommand.
func BetaTestersInviteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("invite", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	email := fs.String("email", "", "Tester email address")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "invite",
		ShortUsage: "asc beta-testers invite [flags]",
		ShortHelp:  "Invite a TestFlight beta tester.",
		LongHelp: `Invite a TestFlight beta tester.

Examples:
  asc beta-testers invite --app "APP_ID" --email "tester@example.com"`,
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
			testerID, err := findBetaTesterIDByEmail(requestCtx, client, resolvedAppID, emailValue)
			if err != nil {
				if !errors.Is(err, errBetaTesterNotFound) {
					return fmt.Errorf("beta-testers invite: %w", err)
				}

				created, createErr := client.CreateBetaTester(requestCtx, emailValue, "", "", nil)
				if createErr != nil {
					return fmt.Errorf("beta-testers invite: failed to create tester: %w", createErr)
				}
				testerID = created.Data.ID
			}

			invitation, err := client.CreateBetaTesterInvitation(requestCtx, resolvedAppID, testerID)
			if err != nil {
				return fmt.Errorf("beta-testers invite: failed to create invitation: %w", err)
			}

			result := asc.BetaTesterInvitationResult{
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
