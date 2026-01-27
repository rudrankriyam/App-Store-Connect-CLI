package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// XcodeCloudCommand returns the xcode-cloud command with subcommands.
func XcodeCloudCommand() *ffcli.Command {
	fs := flag.NewFlagSet("xcode-cloud", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "xcode-cloud",
		ShortUsage: "asc xcode-cloud <subcommand> [flags]",
		ShortHelp:  "Trigger and monitor Xcode Cloud workflows.",
		LongHelp: `Trigger and monitor Xcode Cloud workflows.

Examples:
  asc xcode-cloud workflows --app "APP_ID"
  asc xcode-cloud build-runs --workflow-id "WORKFLOW_ID"
  asc xcode-cloud actions --run-id "BUILD_RUN_ID"
  asc xcode-cloud run --app "APP_ID" --workflow "WorkflowName" --branch "main"
  asc xcode-cloud run --workflow-id "WORKFLOW_ID" --git-reference-id "REF_ID"
  asc xcode-cloud run --app "APP_ID" --workflow "Deploy" --branch "main" --wait
  asc xcode-cloud status --run-id "BUILD_RUN_ID"
  asc xcode-cloud status --run-id "BUILD_RUN_ID" --wait`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			XcodeCloudRunCommand(),
			XcodeCloudStatusCommand(),
			XcodeCloudWorkflowsCommand(),
			XcodeCloudBuildRunsCommand(),
			XcodeCloudActionsCommand(),
			XcodeCloudArtifactsCommand(),
			XcodeCloudTestResultsCommand(),
			XcodeCloudIssuesCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// XcodeCloudRunCommand returns the xcode-cloud run subcommand.
func XcodeCloudRunCommand() *ffcli.Command {
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	workflowName := fs.String("workflow", "", "Workflow name to trigger")
	workflowID := fs.String("workflow-id", "", "Workflow ID to trigger (alternative to --workflow)")
	branch := fs.String("branch", "", "Branch or tag name to build")
	gitReferenceID := fs.String("git-reference-id", "", "Git reference ID to build (alternative to --branch)")
	wait := fs.Bool("wait", false, "Wait for build to complete")
	pollInterval := fs.Duration("poll-interval", 10*time.Second, "Poll interval when waiting")
	timeout := fs.Duration("timeout", 0, "Timeout for Xcode Cloud requests (0 = use ASC_TIMEOUT or 30m default)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "run",
		ShortUsage: "asc xcode-cloud run [flags]",
		ShortHelp:  "Trigger an Xcode Cloud workflow build.",
		LongHelp: `Trigger an Xcode Cloud workflow build.

You can specify the workflow by name (requires --app) or by ID (--workflow-id).
You can specify the branch/tag by name (--branch) or by ID (--git-reference-id).

Examples:
  asc xcode-cloud run --app "123456789" --workflow "CI" --branch "main"
  asc xcode-cloud run --workflow-id "WORKFLOW_ID" --git-reference-id "REF_ID"
  asc xcode-cloud run --app "123456789" --workflow "Deploy" --branch "release/1.0" --wait
  asc xcode-cloud run --app "123456789" --workflow "CI" --branch "main" --wait --poll-interval 30s --timeout 1h`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			// Validate input combinations
			hasWorkflowName := strings.TrimSpace(*workflowName) != ""
			hasWorkflowID := strings.TrimSpace(*workflowID) != ""
			hasBranch := strings.TrimSpace(*branch) != ""
			hasGitRefID := strings.TrimSpace(*gitReferenceID) != ""

			if hasWorkflowName && hasWorkflowID {
				return fmt.Errorf("xcode-cloud run: --workflow and --workflow-id are mutually exclusive")
			}
			if !hasWorkflowName && !hasWorkflowID {
				fmt.Fprintln(os.Stderr, "Error: --workflow or --workflow-id is required")
				return flag.ErrHelp
			}
			if hasBranch && hasGitRefID {
				return fmt.Errorf("xcode-cloud run: --branch and --git-reference-id are mutually exclusive")
			}
			if !hasBranch && !hasGitRefID {
				fmt.Fprintln(os.Stderr, "Error: --branch or --git-reference-id is required")
				return flag.ErrHelp
			}
			if *timeout < 0 {
				return fmt.Errorf("xcode-cloud run: --timeout must be greater than or equal to 0")
			}
			if *wait && *pollInterval <= 0 {
				return fmt.Errorf("xcode-cloud run: --poll-interval must be greater than 0")
			}

			resolvedAppID := resolveAppID(*appID)
			if hasWorkflowName && resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required when using --workflow (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud run: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, *timeout)
			defer cancel()

			// Resolve workflow ID
			resolvedWorkflowID := strings.TrimSpace(*workflowID)
			var workflowNameForOutput string
			if resolvedWorkflowID == "" {
				// Need to resolve workflow by name
				product, err := client.ResolveCiProductForApp(requestCtx, resolvedAppID)
				if err != nil {
					return fmt.Errorf("xcode-cloud run: %w", err)
				}

				workflow, err := client.ResolveCiWorkflowByName(requestCtx, product.ID, strings.TrimSpace(*workflowName))
				if err != nil {
					return fmt.Errorf("xcode-cloud run: %w", err)
				}

				resolvedWorkflowID = workflow.ID
				workflowNameForOutput = workflow.Attributes.Name
			}

			// Resolve git reference ID
			resolvedGitRefID := strings.TrimSpace(*gitReferenceID)
			var refNameForOutput string
			if resolvedGitRefID == "" {
				// Need to resolve git reference by name
				// First get the repository from the workflow
				repo, err := client.GetCiWorkflowRepository(requestCtx, resolvedWorkflowID)
				if err != nil {
					return fmt.Errorf("xcode-cloud run: failed to get workflow repository: %w", err)
				}

				gitRef, err := client.ResolveGitReferenceByName(requestCtx, repo.ID, strings.TrimSpace(*branch))
				if err != nil {
					return fmt.Errorf("xcode-cloud run: %w", err)
				}

				resolvedGitRefID = gitRef.ID
				refNameForOutput = gitRef.Attributes.Name
			}

			// Create the build run
			req := asc.CiBuildRunCreateRequest{
				Data: asc.CiBuildRunCreateData{
					Type: asc.ResourceTypeCiBuildRuns,
					Relationships: &asc.CiBuildRunCreateRelationships{
						Workflow: &asc.Relationship{
							Data: asc.ResourceData{Type: asc.ResourceTypeCiWorkflows, ID: resolvedWorkflowID},
						},
						SourceBranchOrTag: &asc.Relationship{
							Data: asc.ResourceData{Type: asc.ResourceTypeScmGitReferences, ID: resolvedGitRefID},
						},
					},
				},
			}

			resp, err := client.CreateCiBuildRun(requestCtx, req)
			if err != nil {
				return fmt.Errorf("xcode-cloud run: failed to trigger build: %w", err)
			}

			result := &asc.XcodeCloudRunResult{
				BuildRunID:        resp.Data.ID,
				BuildNumber:       resp.Data.Attributes.Number,
				WorkflowID:        resolvedWorkflowID,
				WorkflowName:      workflowNameForOutput,
				GitReferenceID:    resolvedGitRefID,
				GitReferenceName:  refNameForOutput,
				ExecutionProgress: string(resp.Data.Attributes.ExecutionProgress),
				CompletionStatus:  string(resp.Data.Attributes.CompletionStatus),
				StartReason:       resp.Data.Attributes.StartReason,
				CreatedDate:       resp.Data.Attributes.CreatedDate,
				StartedDate:       resp.Data.Attributes.StartedDate,
				FinishedDate:      resp.Data.Attributes.FinishedDate,
			}

			if !*wait {
				return printOutput(result, *output, *pretty)
			}

			// Wait for completion
			return waitForBuildCompletion(requestCtx, client, resp.Data.ID, *pollInterval, *output, *pretty)
		},
	}
}

// XcodeCloudStatusCommand returns the xcode-cloud status subcommand.
func XcodeCloudStatusCommand() *ffcli.Command {
	fs := flag.NewFlagSet("status", flag.ExitOnError)

	runID := fs.String("run-id", "", "Build run ID to check")
	wait := fs.Bool("wait", false, "Wait for build to complete")
	pollInterval := fs.Duration("poll-interval", 10*time.Second, "Poll interval when waiting")
	timeout := fs.Duration("timeout", 0, "Timeout for Xcode Cloud requests (0 = use ASC_TIMEOUT or 30m default)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "status",
		ShortUsage: "asc xcode-cloud status [flags]",
		ShortHelp:  "Check the status of an Xcode Cloud build run.",
		LongHelp: `Check the status of an Xcode Cloud build run.

Examples:
  asc xcode-cloud status --run-id "BUILD_RUN_ID"
  asc xcode-cloud status --run-id "BUILD_RUN_ID" --output table
  asc xcode-cloud status --run-id "BUILD_RUN_ID" --wait
  asc xcode-cloud status --run-id "BUILD_RUN_ID" --wait --poll-interval 30s --timeout 1h`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*runID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --run-id is required")
				return flag.ErrHelp
			}
			if *timeout < 0 {
				return fmt.Errorf("xcode-cloud status: --timeout must be greater than or equal to 0")
			}
			if *wait && *pollInterval <= 0 {
				return fmt.Errorf("xcode-cloud status: --poll-interval must be greater than 0")
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud status: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, *timeout)
			defer cancel()

			if *wait {
				return waitForBuildCompletion(requestCtx, client, strings.TrimSpace(*runID), *pollInterval, *output, *pretty)
			}

			// Single status check
			resp, err := getCiBuildRunWithRetry(requestCtx, client, strings.TrimSpace(*runID))
			if err != nil {
				return fmt.Errorf("xcode-cloud status: %w", err)
			}

			result := buildStatusResult(resp)
			return printOutput(result, *output, *pretty)
		},
	}
}

// XcodeCloudWorkflowsCommand returns the xcode-cloud workflows subcommand.
func XcodeCloudWorkflowsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("workflows", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "workflows",
		ShortUsage: "asc xcode-cloud workflows [flags]",
		ShortHelp:  "List Xcode Cloud workflows for an app.",
		LongHelp: `List Xcode Cloud workflows for an app.

Examples:
  asc xcode-cloud workflows --app "APP_ID"
  asc xcode-cloud workflows --app "APP_ID" --limit 50
  asc xcode-cloud workflows --app "APP_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("xcode-cloud workflows: --limit must be between 1 and 200")
			}
			nextURL := strings.TrimSpace(*next)
			if err := validateNextURL(nextURL); err != nil {
				return fmt.Errorf("xcode-cloud workflows: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && nextURL == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud workflows: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			productID := ""
			if nextURL == "" && resolvedAppID != "" {
				product, err := client.ResolveCiProductForApp(requestCtx, resolvedAppID)
				if err != nil {
					return fmt.Errorf("xcode-cloud workflows: %w", err)
				}
				productID = product.ID
			}

			opts := []asc.CiWorkflowsOption{
				asc.WithCiWorkflowsLimit(*limit),
				asc.WithCiWorkflowsNextURL(nextURL),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithCiWorkflowsLimit(200))
				firstPage, err := client.GetCiWorkflows(requestCtx, productID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("xcode-cloud workflows: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetCiWorkflows(ctx, productID, asc.WithCiWorkflowsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("xcode-cloud workflows: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetCiWorkflows(requestCtx, productID, opts...)
			if err != nil {
				return fmt.Errorf("xcode-cloud workflows: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudBuildRunsCommand returns the xcode-cloud build-runs subcommand.
func XcodeCloudBuildRunsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("build-runs", flag.ExitOnError)

	workflowID := fs.String("workflow-id", "", "Workflow ID to list build runs for")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "build-runs",
		ShortUsage: "asc xcode-cloud build-runs [flags]",
		ShortHelp:  "List Xcode Cloud build runs for a workflow.",
		LongHelp: `List Xcode Cloud build runs for a workflow.

Examples:
  asc xcode-cloud build-runs --workflow-id "WORKFLOW_ID"
  asc xcode-cloud build-runs --workflow-id "WORKFLOW_ID" --limit 50
  asc xcode-cloud build-runs --workflow-id "WORKFLOW_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("xcode-cloud build-runs: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("xcode-cloud build-runs: %w", err)
			}

			resolvedWorkflowID := strings.TrimSpace(*workflowID)
			if resolvedWorkflowID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --workflow-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud build-runs: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			opts := []asc.CiBuildRunsOption{
				asc.WithCiBuildRunsLimit(*limit),
				asc.WithCiBuildRunsNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithCiBuildRunsLimit(200))
				firstPage, err := client.GetCiBuildRuns(requestCtx, resolvedWorkflowID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("xcode-cloud build-runs: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetCiBuildRuns(ctx, resolvedWorkflowID, asc.WithCiBuildRunsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("xcode-cloud build-runs: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetCiBuildRuns(requestCtx, resolvedWorkflowID, opts...)
			if err != nil {
				return fmt.Errorf("xcode-cloud build-runs: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudActionsCommand returns the xcode-cloud actions subcommand.
func XcodeCloudActionsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("actions", flag.ExitOnError)

	runID := fs.String("run-id", "", "Build run ID to get actions for (required)")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "actions",
		ShortUsage: "asc xcode-cloud actions [flags]",
		ShortHelp:  "List build actions for an Xcode Cloud build run.",
		LongHelp: `List build actions for an Xcode Cloud build run.

Build actions show the individual steps of a build run (e.g., "Resolve Dependencies",
"Archive", "Upload") and their status, which helps diagnose why builds failed.

Examples:
  asc xcode-cloud actions --run-id "BUILD_RUN_ID"
  asc xcode-cloud actions --run-id "BUILD_RUN_ID" --output table
  asc xcode-cloud actions --run-id "BUILD_RUN_ID" --limit 50
  asc xcode-cloud actions --run-id "BUILD_RUN_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("xcode-cloud actions: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("xcode-cloud actions: %w", err)
			}

			resolvedRunID := strings.TrimSpace(*runID)
			if resolvedRunID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --run-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud actions: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			opts := []asc.CiBuildActionsOption{
				asc.WithCiBuildActionsLimit(*limit),
				asc.WithCiBuildActionsNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithCiBuildActionsLimit(200))
				firstPage, err := client.GetCiBuildActions(requestCtx, resolvedRunID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("xcode-cloud actions: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetCiBuildActions(ctx, resolvedRunID, asc.WithCiBuildActionsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("xcode-cloud actions: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetCiBuildActions(requestCtx, resolvedRunID, opts...)
			if err != nil {
				return fmt.Errorf("xcode-cloud actions: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudArtifactsCommand returns the xcode-cloud artifacts command with subcommands.
func XcodeCloudArtifactsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("artifacts", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "artifacts",
		ShortUsage: "asc xcode-cloud artifacts <subcommand> [flags]",
		ShortHelp:  "Manage Xcode Cloud build artifacts.",
		LongHelp: `Manage Xcode Cloud build artifacts.

Examples:
  asc xcode-cloud artifacts list --action-id "ACTION_ID"
  asc xcode-cloud artifacts get --id "ARTIFACT_ID"
  asc xcode-cloud artifacts download --id "ARTIFACT_ID" --path ./artifact.zip`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			XcodeCloudArtifactsListCommand(),
			XcodeCloudArtifactsGetCommand(),
			XcodeCloudArtifactsDownloadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// XcodeCloudArtifactsListCommand returns the xcode-cloud artifacts list subcommand.
func XcodeCloudArtifactsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	actionID := fs.String("action-id", "", "Build action ID to list artifacts for")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc xcode-cloud artifacts list [flags]",
		ShortHelp:  "List artifacts for a build action.",
		LongHelp: `List artifacts for a build action.

Examples:
  asc xcode-cloud artifacts list --action-id "ACTION_ID"
  asc xcode-cloud artifacts list --action-id "ACTION_ID" --output table
  asc xcode-cloud artifacts list --action-id "ACTION_ID" --limit 50
  asc xcode-cloud artifacts list --action-id "ACTION_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("xcode-cloud artifacts list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("xcode-cloud artifacts list: %w", err)
			}

			resolvedActionID := strings.TrimSpace(*actionID)
			if resolvedActionID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --action-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts list: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			opts := []asc.CiArtifactsOption{
				asc.WithCiArtifactsLimit(*limit),
				asc.WithCiArtifactsNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithCiArtifactsLimit(200))
				firstPage, err := client.GetCiBuildActionArtifacts(requestCtx, resolvedActionID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("xcode-cloud artifacts list: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetCiBuildActionArtifacts(ctx, resolvedActionID, asc.WithCiArtifactsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("xcode-cloud artifacts list: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetCiBuildActionArtifacts(requestCtx, resolvedActionID, opts...)
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts list: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudArtifactsGetCommand returns the xcode-cloud artifacts get subcommand.
func XcodeCloudArtifactsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	id := fs.String("id", "", "Artifact ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc xcode-cloud artifacts get --id \"ARTIFACT_ID\"",
		ShortHelp:  "Get details for a build artifact.",
		LongHelp: `Get details for a build artifact.

Examples:
  asc xcode-cloud artifacts get --id "ARTIFACT_ID"
  asc xcode-cloud artifacts get --id "ARTIFACT_ID" --output table`,
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
				return fmt.Errorf("xcode-cloud artifacts get: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			resp, err := client.GetCiArtifact(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts get: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudArtifactsDownloadCommand returns the xcode-cloud artifacts download subcommand.
func XcodeCloudArtifactsDownloadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("download", flag.ExitOnError)

	id := fs.String("id", "", "Artifact ID")
	path := fs.String("path", "", "Output file path for the artifact")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing file")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "download",
		ShortUsage: "asc xcode-cloud artifacts download --id \"ARTIFACT_ID\" --path ./artifact.zip",
		ShortHelp:  "Download a build artifact.",
		LongHelp: `Download a build artifact.

Examples:
  asc xcode-cloud artifacts download --id "ARTIFACT_ID" --path ./artifact.zip
  asc xcode-cloud artifacts download --id "ARTIFACT_ID" --path ./artifact.zip --overwrite`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}
			pathValue := strings.TrimSpace(*path)
			if pathValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --path is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts download: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			artifactResp, err := client.GetCiArtifact(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts download: failed to fetch artifact: %w", err)
			}

			downloadURL := strings.TrimSpace(artifactResp.Data.Attributes.DownloadURL)
			if downloadURL == "" {
				return fmt.Errorf("xcode-cloud artifacts download: artifact has no download URL")
			}

			download, err := client.DownloadCiArtifact(requestCtx, downloadURL)
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts download: %w", err)
			}
			defer download.Body.Close()

			bytesWritten, err := writeArtifactFile(pathValue, download.Body, *overwrite)
			if err != nil {
				return fmt.Errorf("xcode-cloud artifacts download: %w", err)
			}

			result := &asc.CiArtifactDownloadResult{
				ID:           artifactResp.Data.ID,
				FileName:     artifactResp.Data.Attributes.FileName,
				FileType:     artifactResp.Data.Attributes.FileType,
				FileSize:     artifactResp.Data.Attributes.FileSize,
				OutputPath:   pathValue,
				BytesWritten: bytesWritten,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

// XcodeCloudTestResultsCommand returns the xcode-cloud test-results command with subcommands.
func XcodeCloudTestResultsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("test-results", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "test-results",
		ShortUsage: "asc xcode-cloud test-results <subcommand> [flags]",
		ShortHelp:  "List Xcode Cloud test results.",
		LongHelp: `List Xcode Cloud test results.

Examples:
  asc xcode-cloud test-results list --action-id "ACTION_ID"
  asc xcode-cloud test-results get --id "TEST_RESULT_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			XcodeCloudTestResultsListCommand(),
			XcodeCloudTestResultsGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// XcodeCloudTestResultsListCommand returns the xcode-cloud test-results list subcommand.
func XcodeCloudTestResultsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	actionID := fs.String("action-id", "", "Build action ID to list test results for")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc xcode-cloud test-results list [flags]",
		ShortHelp:  "List test results for a build action.",
		LongHelp: `List test results for a build action.

Examples:
  asc xcode-cloud test-results list --action-id "ACTION_ID"
  asc xcode-cloud test-results list --action-id "ACTION_ID" --output table
  asc xcode-cloud test-results list --action-id "ACTION_ID" --limit 50
  asc xcode-cloud test-results list --action-id "ACTION_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("xcode-cloud test-results list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("xcode-cloud test-results list: %w", err)
			}

			resolvedActionID := strings.TrimSpace(*actionID)
			if resolvedActionID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --action-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud test-results list: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			opts := []asc.CiTestResultsOption{
				asc.WithCiTestResultsLimit(*limit),
				asc.WithCiTestResultsNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithCiTestResultsLimit(200))
				firstPage, err := client.GetCiBuildActionTestResults(requestCtx, resolvedActionID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("xcode-cloud test-results list: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetCiBuildActionTestResults(ctx, resolvedActionID, asc.WithCiTestResultsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("xcode-cloud test-results list: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetCiBuildActionTestResults(requestCtx, resolvedActionID, opts...)
			if err != nil {
				return fmt.Errorf("xcode-cloud test-results list: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudTestResultsGetCommand returns the xcode-cloud test-results get subcommand.
func XcodeCloudTestResultsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	id := fs.String("id", "", "Test result ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc xcode-cloud test-results get --id \"TEST_RESULT_ID\"",
		ShortHelp:  "Get details for a test result.",
		LongHelp: `Get details for a test result.

Examples:
  asc xcode-cloud test-results get --id "TEST_RESULT_ID"
  asc xcode-cloud test-results get --id "TEST_RESULT_ID" --output table`,
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
				return fmt.Errorf("xcode-cloud test-results get: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			resp, err := client.GetCiTestResult(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("xcode-cloud test-results get: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudIssuesCommand returns the xcode-cloud issues command with subcommands.
func XcodeCloudIssuesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("issues", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "issues",
		ShortUsage: "asc xcode-cloud issues <subcommand> [flags]",
		ShortHelp:  "List Xcode Cloud build issues.",
		LongHelp: `List Xcode Cloud build issues.

Examples:
  asc xcode-cloud issues list --action-id "ACTION_ID"
  asc xcode-cloud issues get --id "ISSUE_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			XcodeCloudIssuesListCommand(),
			XcodeCloudIssuesGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// XcodeCloudIssuesListCommand returns the xcode-cloud issues list subcommand.
func XcodeCloudIssuesListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	actionID := fs.String("action-id", "", "Build action ID to list issues for")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc xcode-cloud issues list [flags]",
		ShortHelp:  "List issues for a build action.",
		LongHelp: `List issues for a build action.

Examples:
  asc xcode-cloud issues list --action-id "ACTION_ID"
  asc xcode-cloud issues list --action-id "ACTION_ID" --output table
  asc xcode-cloud issues list --action-id "ACTION_ID" --limit 50
  asc xcode-cloud issues list --action-id "ACTION_ID" --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("xcode-cloud issues list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("xcode-cloud issues list: %w", err)
			}

			resolvedActionID := strings.TrimSpace(*actionID)
			if resolvedActionID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --action-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("xcode-cloud issues list: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			opts := []asc.CiIssuesOption{
				asc.WithCiIssuesLimit(*limit),
				asc.WithCiIssuesNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithCiIssuesLimit(200))
				firstPage, err := client.GetCiBuildActionIssues(requestCtx, resolvedActionID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("xcode-cloud issues list: failed to fetch: %w", err)
				}

				resp, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetCiBuildActionIssues(ctx, resolvedActionID, asc.WithCiIssuesNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("xcode-cloud issues list: %w", err)
				}

				return printOutput(resp, *output, *pretty)
			}

			resp, err := client.GetCiBuildActionIssues(requestCtx, resolvedActionID, opts...)
			if err != nil {
				return fmt.Errorf("xcode-cloud issues list: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// XcodeCloudIssuesGetCommand returns the xcode-cloud issues get subcommand.
func XcodeCloudIssuesGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	id := fs.String("id", "", "Issue ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc xcode-cloud issues get --id \"ISSUE_ID\"",
		ShortHelp:  "Get details for a build issue.",
		LongHelp: `Get details for a build issue.

Examples:
  asc xcode-cloud issues get --id "ISSUE_ID"
  asc xcode-cloud issues get --id "ISSUE_ID" --output table`,
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
				return fmt.Errorf("xcode-cloud issues get: %w", err)
			}

			requestCtx, cancel := contextWithXcodeCloudTimeout(ctx, 0)
			defer cancel()

			resp, err := client.GetCiIssue(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("xcode-cloud issues get: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// waitForBuildCompletion polls until the build run completes or times out.
func waitForBuildCompletion(ctx context.Context, client *asc.Client, buildRunID string, pollInterval time.Duration, outputFormat string, pretty bool) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		resp, err := getCiBuildRunWithRetry(ctx, client, buildRunID)
		if err != nil {
			return fmt.Errorf("xcode-cloud: failed to check status: %w", err)
		}

		if asc.IsBuildRunComplete(resp.Data.Attributes.ExecutionProgress) {
			result := buildStatusResult(resp)
			if err := printOutput(result, outputFormat, pretty); err != nil {
				return err
			}

			// Return error for failed builds
			if !asc.IsBuildRunSuccessful(resp.Data.Attributes.CompletionStatus) {
				return fmt.Errorf("build run %s completed with status: %s", buildRunID, resp.Data.Attributes.CompletionStatus)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return fmt.Errorf("xcode-cloud: canceled waiting for build run %s (last status: %s)", buildRunID, resp.Data.Attributes.ExecutionProgress)
			}
			return fmt.Errorf("xcode-cloud: timed out waiting for build run %s (last status: %s)", buildRunID, resp.Data.Attributes.ExecutionProgress)
		case <-ticker.C:
			// Continue polling
		}
	}
}

// buildStatusResult converts a CiBuildRunResponse to XcodeCloudStatusResult.
func buildStatusResult(resp *asc.CiBuildRunResponse) *asc.XcodeCloudStatusResult {
	result := &asc.XcodeCloudStatusResult{
		BuildRunID:        resp.Data.ID,
		BuildNumber:       resp.Data.Attributes.Number,
		ExecutionProgress: string(resp.Data.Attributes.ExecutionProgress),
		CompletionStatus:  string(resp.Data.Attributes.CompletionStatus),
		StartReason:       resp.Data.Attributes.StartReason,
		CancelReason:      resp.Data.Attributes.CancelReason,
		CreatedDate:       resp.Data.Attributes.CreatedDate,
		StartedDate:       resp.Data.Attributes.StartedDate,
		FinishedDate:      resp.Data.Attributes.FinishedDate,
		SourceCommit:      resp.Data.Attributes.SourceCommit,
		IssueCounts:       resp.Data.Attributes.IssueCounts,
	}

	if resp.Data.Relationships != nil && resp.Data.Relationships.Workflow != nil {
		result.WorkflowID = resp.Data.Relationships.Workflow.Data.ID
	}

	return result
}

func writeArtifactFile(path string, reader io.Reader, overwrite bool) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}

	if !overwrite {
		file, err := openNewFileNoFollow(path, 0o600)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				return 0, fmt.Errorf("output file already exists: %w", err)
			}
			return 0, err
		}
		defer file.Close()

		n, err := io.Copy(file, reader)
		if err != nil {
			return 0, err
		}
		if err := file.Sync(); err != nil {
			return 0, err
		}
		return n, nil
	}

	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return 0, fmt.Errorf("refusing to overwrite symlink %q", path)
		}
		if info.IsDir() {
			return 0, fmt.Errorf("output path %q is a directory", path)
		}
		if err := os.Remove(path); err != nil {
			return 0, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}

	tempFile, err := os.CreateTemp(filepath.Dir(path), ".asc-artifact-*")
	if err != nil {
		return 0, err
	}
	defer tempFile.Close()

	tempPath := tempFile.Name()
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	n, err := io.Copy(tempFile, reader)
	if err != nil {
		return 0, err
	}
	if err := tempFile.Sync(); err != nil {
		return 0, err
	}
	if err := tempFile.Close(); err != nil {
		return 0, err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return 0, err
	}

	success = true
	return n, nil
}

const defaultXcodeCloudTimeout = 30 * time.Minute

func contextWithXcodeCloudTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = asc.ResolveTimeoutWithDefault(defaultXcodeCloudTimeout)
	}
	return context.WithTimeout(ctx, timeout)
}

func getCiBuildRunWithRetry(ctx context.Context, client *asc.Client, buildRunID string) (*asc.CiBuildRunResponse, error) {
	retryOpts := asc.ResolveRetryOptions()
	return asc.WithRetry(ctx, func() (*asc.CiBuildRunResponse, error) {
		resp, err := client.GetCiBuildRun(ctx, buildRunID)
		if err != nil {
			if isTransientNetworkError(err) {
				return nil, &asc.RetryableError{Err: err}
			}
			return nil, err
		}
		return resp, nil
	}, retryOpts)
}

func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	return errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNREFUSED)
}
