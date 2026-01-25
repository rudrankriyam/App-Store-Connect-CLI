package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
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
