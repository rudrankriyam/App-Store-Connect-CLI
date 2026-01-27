package asc

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"
)

func captureXcodeCloudStdout(t *testing.T, fn func() error) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stdout = w

	err = fn()

	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("close error: %v", closeErr)
	}
	os.Stdout = orig

	var buf bytes.Buffer
	if _, readErr := io.Copy(&buf, r); readErr != nil {
		t.Fatalf("read error: %v", readErr)
	}
	if err != nil {
		t.Fatalf("function error: %v", err)
	}

	return buf.String()
}

func TestPrintTable_XcodeCloudRunResult(t *testing.T) {
	result := &XcodeCloudRunResult{
		BuildRunID:        "run-123",
		BuildNumber:       42,
		WorkflowID:        "wf-456",
		WorkflowName:      "CI Build",
		GitReferenceID:    "ref-789",
		GitReferenceName:  "main",
		ExecutionProgress: "PENDING",
		CompletionStatus:  "",
		StartReason:       "MANUAL",
		CreatedDate:       "2026-01-22T10:00:00Z",
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(result)
	})

	if !strings.Contains(output, "Build Run ID") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "run-123") {
		t.Fatalf("expected build run ID in output, got: %s", output)
	}
	if !strings.Contains(output, "CI Build") {
		t.Fatalf("expected workflow name in output, got: %s", output)
	}
	if !strings.Contains(output, "PENDING") {
		t.Fatalf("expected execution progress in output, got: %s", output)
	}
}

func TestPrintMarkdown_XcodeCloudRunResult(t *testing.T) {
	result := &XcodeCloudRunResult{
		BuildRunID:        "run-123",
		BuildNumber:       42,
		WorkflowID:        "wf-456",
		WorkflowName:      "CI Build",
		GitReferenceID:    "ref-789",
		GitReferenceName:  "main",
		ExecutionProgress: "RUNNING",
		CompletionStatus:  "",
		StartReason:       "MANUAL",
		CreatedDate:       "2026-01-22T10:00:00Z",
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(result)
	})

	if !strings.Contains(output, "| Build Run ID |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "run-123") {
		t.Fatalf("expected build run ID in output, got: %s", output)
	}
	if !strings.Contains(output, "RUNNING") {
		t.Fatalf("expected execution progress in output, got: %s", output)
	}
}

func TestPrintTable_XcodeCloudStatusResult(t *testing.T) {
	result := &XcodeCloudStatusResult{
		BuildRunID:        "run-123",
		BuildNumber:       42,
		WorkflowID:        "wf-456",
		ExecutionProgress: "COMPLETE",
		CompletionStatus:  "SUCCEEDED",
		StartReason:       "MANUAL",
		CancelReason:      "",
		CreatedDate:       "2026-01-22T10:00:00Z",
		StartedDate:       "2026-01-22T10:01:00Z",
		FinishedDate:      "2026-01-22T10:05:00Z",
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(result)
	})

	if !strings.Contains(output, "Build Run ID") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "COMPLETE") {
		t.Fatalf("expected execution progress in output, got: %s", output)
	}
	if !strings.Contains(output, "SUCCEEDED") {
		t.Fatalf("expected completion status in output, got: %s", output)
	}
}

func TestPrintMarkdown_XcodeCloudStatusResult(t *testing.T) {
	result := &XcodeCloudStatusResult{
		BuildRunID:        "run-123",
		BuildNumber:       42,
		WorkflowID:        "wf-456",
		ExecutionProgress: "COMPLETE",
		CompletionStatus:  "FAILED",
		StartReason:       "MANUAL",
		CancelReason:      "",
		CreatedDate:       "2026-01-22T10:00:00Z",
		StartedDate:       "2026-01-22T10:01:00Z",
		FinishedDate:      "2026-01-22T10:05:00Z",
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(result)
	})

	if !strings.Contains(output, "| Build Run ID |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "FAILED") {
		t.Fatalf("expected completion status in output, got: %s", output)
	}
}

func TestPrintTable_CiProducts(t *testing.T) {
	resp := &CiProductsResponse{
		Data: []CiProductResource{
			{
				ID: "prod-1",
				Attributes: CiProductAttributes{
					Name:        "MyApp",
					BundleID:    "com.example.myapp",
					ProductType: "APP",
					CreatedDate: "2026-01-22T10:00:00Z",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Bundle ID") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "com.example.myapp") {
		t.Fatalf("expected bundle ID in output, got: %s", output)
	}
}

func TestPrintMarkdown_CiProducts(t *testing.T) {
	resp := &CiProductsResponse{
		Data: []CiProductResource{
			{
				ID: "prod-1",
				Attributes: CiProductAttributes{
					Name:        "MyApp",
					BundleID:    "com.example.myapp",
					ProductType: "APP",
					CreatedDate: "2026-01-22T10:00:00Z",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Name | Bundle ID |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "MyApp") {
		t.Fatalf("expected app name in output, got: %s", output)
	}
}

func TestPrintTable_CiWorkflows(t *testing.T) {
	resp := &CiWorkflowsResponse{
		Data: []CiWorkflowResource{
			{
				ID: "wf-1",
				Attributes: CiWorkflowAttributes{
					Name:             "CI Build",
					IsEnabled:        true,
					LastModifiedDate: "2026-01-22T10:00:00Z",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Enabled") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "CI Build") {
		t.Fatalf("expected workflow name in output, got: %s", output)
	}
}

func TestPrintMarkdown_CiWorkflows(t *testing.T) {
	resp := &CiWorkflowsResponse{
		Data: []CiWorkflowResource{
			{
				ID: "wf-1",
				Attributes: CiWorkflowAttributes{
					Name:             "Deploy",
					IsEnabled:        false,
					LastModifiedDate: "2026-01-22T10:00:00Z",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Name | Enabled |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "Deploy") {
		t.Fatalf("expected workflow name in output, got: %s", output)
	}
}

func TestPrintTable_CiBuildRuns(t *testing.T) {
	resp := &CiBuildRunsResponse{
		Data: []CiBuildRunResource{
			{
				ID: "run-1",
				Attributes: CiBuildRunAttributes{
					Number:            1,
					ExecutionProgress: CiBuildRunExecutionProgressComplete,
					CompletionStatus:  CiBuildRunCompletionStatusSucceeded,
					StartReason:       "MANUAL",
					CreatedDate:       "2026-01-22T10:00:00Z",
					StartedDate:       "2026-01-22T10:01:00Z",
					FinishedDate:      "2026-01-22T10:05:00Z",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Progress") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "COMPLETE") {
		t.Fatalf("expected execution progress in output, got: %s", output)
	}
}

func TestPrintMarkdown_CiBuildRuns(t *testing.T) {
	resp := &CiBuildRunsResponse{
		Data: []CiBuildRunResource{
			{
				ID: "run-1",
				Attributes: CiBuildRunAttributes{
					Number:            1,
					ExecutionProgress: CiBuildRunExecutionProgressRunning,
					CompletionStatus:  "",
					StartReason:       "MANUAL",
					CreatedDate:       "2026-01-22T10:00:00Z",
					StartedDate:       "2026-01-22T10:01:00Z",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Build # |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "RUNNING") {
		t.Fatalf("expected execution progress in output, got: %s", output)
	}
}

func TestPrintTable_CiArtifacts(t *testing.T) {
	resp := &CiArtifactsResponse{
		Data: []CiArtifactResource{
			{
				ID: "art-1",
				Attributes: CiArtifactAttributes{
					FileName:    "Build.zip",
					FileType:    "ARCHIVE",
					FileSize:    2048,
					DownloadURL: "https://example.com/artifact.zip",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Download URL") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "Build.zip") {
		t.Fatalf("expected file name in output, got: %s", output)
	}
}

func TestPrintMarkdown_CiArtifacts(t *testing.T) {
	resp := &CiArtifactsResponse{
		Data: []CiArtifactResource{
			{
				ID: "art-1",
				Attributes: CiArtifactAttributes{
					FileName: "Logs.zip",
					FileType: "LOG_BUNDLE",
					FileSize: 512,
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Name | Type |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "Logs.zip") {
		t.Fatalf("expected file name in output, got: %s", output)
	}
}

func TestPrintTable_CiTestResults(t *testing.T) {
	resp := &CiTestResultsResponse{
		Data: []CiTestResultResource{
			{
				ID: "test-1",
				Attributes: CiTestResultAttributes{
					ClassName: "Tests",
					Name:      "testExample",
					Status:    CiTestStatusSuccess,
					DestinationTestResults: []CiTestDestinationResult{
						{Duration: 1.234},
					},
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Duration") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "1.23s") {
		t.Fatalf("expected duration in output, got: %s", output)
	}
}

func TestPrintMarkdown_CiTestResults(t *testing.T) {
	resp := &CiTestResultsResponse{
		Data: []CiTestResultResource{
			{
				ID: "test-1",
				Attributes: CiTestResultAttributes{
					ClassName: "Tests",
					Name:      "testExample",
					Status:    CiTestStatusFailure,
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Class | Name |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "FAILURE") {
		t.Fatalf("expected status in output, got: %s", output)
	}
}

func TestPrintTable_CiIssues(t *testing.T) {
	resp := &CiIssuesResponse{
		Data: []CiIssueResource{
			{
				ID: "issue-1",
				Attributes: CiIssueAttributes{
					IssueType: "ERROR",
					Message:   "Something broke",
					FileSource: &FileLocation{
						Path:       "Sources/App.swift",
						LineNumber: 42,
					},
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(resp)
	})

	if !strings.Contains(output, "Line") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "Sources/App.swift") {
		t.Fatalf("expected file path in output, got: %s", output)
	}
}

func TestPrintMarkdown_CiIssues(t *testing.T) {
	resp := &CiIssuesResponse{
		Data: []CiIssueResource{
			{
				ID: "issue-1",
				Attributes: CiIssueAttributes{
					IssueType: "WARNING",
					Message:   "Check this",
				},
			},
		},
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintMarkdown(resp)
	})

	if !strings.Contains(output, "| ID | Type | File |") {
		t.Fatalf("expected markdown header in output, got: %s", output)
	}
	if !strings.Contains(output, "WARNING") {
		t.Fatalf("expected issue type in output, got: %s", output)
	}
}

func TestPrintTable_CiArtifactDownloadResult(t *testing.T) {
	result := &CiArtifactDownloadResult{
		ID:           "art-1",
		FileName:     "Build.zip",
		FileType:     "ARCHIVE",
		FileSize:     2048,
		OutputPath:   "/tmp/Build.zip",
		BytesWritten: 2048,
	}

	output := captureXcodeCloudStdout(t, func() error {
		return PrintTable(result)
	})

	if !strings.Contains(output, "Bytes Written") {
		t.Fatalf("expected header in output, got: %s", output)
	}
	if !strings.Contains(output, "/tmp/Build.zip") {
		t.Fatalf("expected output path in output, got: %s", output)
	}
}

func TestIsBuildRunComplete(t *testing.T) {
	tests := []struct {
		progress CiBuildRunExecutionProgress
		want     bool
	}{
		{CiBuildRunExecutionProgressPending, false},
		{CiBuildRunExecutionProgressRunning, false},
		{CiBuildRunExecutionProgressComplete, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.progress), func(t *testing.T) {
			got := IsBuildRunComplete(tt.progress)
			if got != tt.want {
				t.Errorf("IsBuildRunComplete(%s) = %v, want %v", tt.progress, got, tt.want)
			}
		})
	}
}

func TestIsBuildRunSuccessful(t *testing.T) {
	tests := []struct {
		status CiBuildRunCompletionStatus
		want   bool
	}{
		{CiBuildRunCompletionStatusSucceeded, true},
		{CiBuildRunCompletionStatusFailed, false},
		{CiBuildRunCompletionStatusErrored, false},
		{CiBuildRunCompletionStatusCanceled, false},
		{CiBuildRunCompletionStatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := IsBuildRunSuccessful(tt.status)
			if got != tt.want {
				t.Errorf("IsBuildRunSuccessful(%s) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestBuildCiProductsQuery(t *testing.T) {
	query := &ciProductsQuery{}
	WithCiProductsAppID("app-1")(query)
	WithCiProductsLimit(25)(query)

	values, err := url.ParseQuery(buildCiProductsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[app]"); got != "app-1" {
		t.Fatalf("expected filter[app]=app-1, got %q", got)
	}
	if got := values.Get("limit"); got != "25" {
		t.Fatalf("expected limit=25, got %q", got)
	}
}

func TestBuildCiWorkflowsQuery(t *testing.T) {
	query := &ciWorkflowsQuery{}
	WithCiWorkflowsLimit(50)(query)

	values, err := url.ParseQuery(buildCiWorkflowsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "50" {
		t.Fatalf("expected limit=50, got %q", got)
	}
}

func TestBuildScmGitReferencesQuery(t *testing.T) {
	query := &scmGitReferencesQuery{}
	WithScmGitReferencesLimit(100)(query)

	values, err := url.ParseQuery(buildScmGitReferencesQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "100" {
		t.Fatalf("expected limit=100, got %q", got)
	}
}

func TestBuildCiBuildRunsQuery(t *testing.T) {
	query := &ciBuildRunsQuery{}
	WithCiBuildRunsLimit(10)(query)

	values, err := url.ParseQuery(buildCiBuildRunsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
}

func TestBuildCiArtifactsQuery(t *testing.T) {
	query := &ciArtifactsQuery{}
	WithCiArtifactsLimit(25)(query)

	values, err := url.ParseQuery(buildCiArtifactsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "25" {
		t.Fatalf("expected limit=25, got %q", got)
	}
}

func TestBuildCiTestResultsQuery(t *testing.T) {
	query := &ciTestResultsQuery{}
	WithCiTestResultsLimit(30)(query)

	values, err := url.ParseQuery(buildCiTestResultsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "30" {
		t.Fatalf("expected limit=30, got %q", got)
	}
}

func TestBuildCiIssuesQuery(t *testing.T) {
	query := &ciIssuesQuery{}
	WithCiIssuesLimit(35)(query)

	values, err := url.ParseQuery(buildCiIssuesQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "35" {
		t.Fatalf("expected limit=35, got %q", got)
	}
}
