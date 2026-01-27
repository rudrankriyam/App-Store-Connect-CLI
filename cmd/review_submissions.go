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

// ReviewSubmissionsCommand returns the review-submissions parent command.
func ReviewSubmissionsCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "review-submissions",
		ShortUsage: "asc review-submissions <subcommand> [flags]",
		ShortHelp:  "Manage App Store review submissions.",
		LongHelp: `Manage App Store review submissions.

Review submissions can include multiple items: app versions, custom product pages,
in-app events, and experiments.

Examples:
  asc review-submissions list --app "123456789"
  asc review-submissions create --app "123456789" --platform IOS
  asc review-submissions submit --id "SUBMISSION_ID" --confirm

Alias:
  asc review submissions-list --app "123456789"`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			ReviewSubmissionsListCommand(),
			ReviewSubmissionsGetCommand(),
			ReviewSubmissionsCreateCommand(),
			ReviewSubmissionsSubmitCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// ReviewSubmissionsListCommand returns the review-submissions list subcommand.
func ReviewSubmissionsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("review-submissions list", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	platform := fs.String("platform", "", "Filter by platform: IOS, MAC_OS, TV_OS, VISION_OS (comma-separated)")
	state := fs.String("state", "", "Filter by state (comma-separated)")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Next page URL from a previous response")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc review-submissions list [flags]",
		ShortHelp:  "List review submissions for an app.",
		LongHelp: `List review submissions for an app.

Examples:
  asc review-submissions list --app "123456789"
  asc review-submissions list --app "123456789" --platform IOS --state READY_FOR_REVIEW
  asc review-submissions list --app "123456789" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("review-submissions list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("review-submissions list: %w", err)
			}

			platforms, err := normalizeAppStoreVersionPlatforms(splitCSVUpper(*platform))
			if err != nil {
				return fmt.Errorf("review-submissions list: %w", err)
			}
			states := splitCSVUpper(*state)

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("review-submissions list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.ReviewSubmissionsOption{
				asc.WithReviewSubmissionsLimit(*limit),
				asc.WithReviewSubmissionsNextURL(*next),
				asc.WithReviewSubmissionsPlatforms(platforms),
				asc.WithReviewSubmissionsStates(states),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithReviewSubmissionsLimit(200))
				firstPage, err := client.GetReviewSubmissions(requestCtx, resolvedAppID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("review-submissions list: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetReviewSubmissions(ctx, resolvedAppID, asc.WithReviewSubmissionsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("review-submissions list: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetReviewSubmissions(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("review-submissions list: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// ReviewSubmissionsGetCommand returns the review-submissions get subcommand.
func ReviewSubmissionsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("review-submissions get", flag.ExitOnError)

	submissionID := fs.String("id", "", "Review submission ID (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc review-submissions get [flags]",
		ShortHelp:  "Get a review submission by ID.",
		LongHelp: `Get a review submission by ID.

Examples:
  asc review-submissions get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*submissionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("review-submissions get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetReviewSubmission(requestCtx, strings.TrimSpace(*submissionID))
			if err != nil {
				return fmt.Errorf("review-submissions get: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// ReviewSubmissionsCreateCommand returns the review-submissions create subcommand.
func ReviewSubmissionsCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("review-submissions create", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	platform := fs.String("platform", "IOS", "Platform: IOS, MAC_OS, TV_OS, VISION_OS")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc review-submissions create [flags]",
		ShortHelp:  "Create a review submission.",
		LongHelp: `Create a review submission for an app.

Examples:
  asc review-submissions create --app "123456789" --platform IOS`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			normalizedPlatform, err := normalizeSubmitPlatform(*platform)
			if err != nil {
				return fmt.Errorf("review-submissions create: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("review-submissions create: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.CreateReviewSubmission(requestCtx, resolvedAppID, asc.Platform(normalizedPlatform))
			if err != nil {
				return fmt.Errorf("review-submissions create: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// ReviewSubmissionsSubmitCommand returns the review-submissions submit subcommand.
func ReviewSubmissionsSubmitCommand() *ffcli.Command {
	fs := flag.NewFlagSet("review-submissions submit", flag.ExitOnError)

	submissionID := fs.String("id", "", "Review submission ID (required)")
	confirm := fs.Bool("confirm", false, "Confirm submission (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "submit",
		ShortUsage: "asc review-submissions submit [flags]",
		ShortHelp:  "Submit a review submission.",
		LongHelp: `Submit a review submission for review.

Examples:
  asc review-submissions submit --id "SUBMISSION_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to submit")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*submissionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("review-submissions submit: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.SubmitReviewSubmission(requestCtx, strings.TrimSpace(*submissionID))
			if err != nil {
				return fmt.Errorf("review-submissions submit: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// ReviewSubmissionsListNounCommand returns the noun-first review submissions list command.
func ReviewSubmissionsListNounCommand() *ffcli.Command {
	cmd := ReviewSubmissionsListCommand()
	cmd.Name = "submissions-list"
	cmd.ShortUsage = strings.ReplaceAll(cmd.ShortUsage, "review-submissions list", "review submissions-list")
	cmd.LongHelp = strings.ReplaceAll(cmd.LongHelp, "review-submissions list", "review submissions-list")
	return cmd
}

// ReviewSubmissionsGetNounCommand returns the noun-first review submissions get command.
func ReviewSubmissionsGetNounCommand() *ffcli.Command {
	cmd := ReviewSubmissionsGetCommand()
	cmd.Name = "submissions-get"
	cmd.ShortUsage = strings.ReplaceAll(cmd.ShortUsage, "review-submissions get", "review submissions-get")
	cmd.LongHelp = strings.ReplaceAll(cmd.LongHelp, "review-submissions get", "review submissions-get")
	return cmd
}

// ReviewSubmissionsCreateNounCommand returns the noun-first review submissions create command.
func ReviewSubmissionsCreateNounCommand() *ffcli.Command {
	cmd := ReviewSubmissionsCreateCommand()
	cmd.Name = "submissions-create"
	cmd.ShortUsage = strings.ReplaceAll(cmd.ShortUsage, "review-submissions create", "review submissions-create")
	cmd.LongHelp = strings.ReplaceAll(cmd.LongHelp, "review-submissions create", "review submissions-create")
	return cmd
}

// ReviewSubmissionsSubmitNounCommand returns the noun-first review submissions submit command.
func ReviewSubmissionsSubmitNounCommand() *ffcli.Command {
	cmd := ReviewSubmissionsSubmitCommand()
	cmd.Name = "submissions-submit"
	cmd.ShortUsage = strings.ReplaceAll(cmd.ShortUsage, "review-submissions submit", "review submissions-submit")
	cmd.LongHelp = strings.ReplaceAll(cmd.LongHelp, "review-submissions submit", "review submissions-submit")
	return cmd
}
