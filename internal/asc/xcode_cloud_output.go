package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// CiArtifactDownloadResult represents CLI output for artifact downloads.
type CiArtifactDownloadResult struct {
	ID           string `json:"id"`
	FileName     string `json:"fileName,omitempty"`
	FileType     string `json:"fileType,omitempty"`
	FileSize     int    `json:"fileSize,omitempty"`
	OutputPath   string `json:"outputPath"`
	BytesWritten int64  `json:"bytesWritten,omitempty"`
}

func printXcodeCloudRunResultTable(result *XcodeCloudRunResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Build Run ID\tBuild #\tWorkflow ID\tWorkflow Name\tGit Ref ID\tGit Ref Name\tProgress\tStatus\tStart Reason\tCreated")
	fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		result.BuildRunID,
		result.BuildNumber,
		result.WorkflowID,
		result.WorkflowName,
		result.GitReferenceID,
		result.GitReferenceName,
		result.ExecutionProgress,
		result.CompletionStatus,
		result.StartReason,
		result.CreatedDate,
	)
	return w.Flush()
}

func printXcodeCloudRunResultMarkdown(result *XcodeCloudRunResult) error {
	fmt.Fprintln(os.Stdout, "| Build Run ID | Build # | Workflow ID | Workflow Name | Git Ref ID | Git Ref Name | Progress | Status | Start Reason | Created |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.BuildRunID),
		result.BuildNumber,
		escapeMarkdown(result.WorkflowID),
		escapeMarkdown(result.WorkflowName),
		escapeMarkdown(result.GitReferenceID),
		escapeMarkdown(result.GitReferenceName),
		escapeMarkdown(result.ExecutionProgress),
		escapeMarkdown(result.CompletionStatus),
		escapeMarkdown(result.StartReason),
		escapeMarkdown(result.CreatedDate),
	)
	return nil
}

func printXcodeCloudStatusResultTable(result *XcodeCloudStatusResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Build Run ID\tBuild #\tWorkflow ID\tProgress\tStatus\tStart Reason\tCancel Reason\tCreated\tStarted\tFinished")
	fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		result.BuildRunID,
		result.BuildNumber,
		result.WorkflowID,
		result.ExecutionProgress,
		result.CompletionStatus,
		result.StartReason,
		result.CancelReason,
		result.CreatedDate,
		result.StartedDate,
		result.FinishedDate,
	)
	return w.Flush()
}

func printXcodeCloudStatusResultMarkdown(result *XcodeCloudStatusResult) error {
	fmt.Fprintln(os.Stdout, "| Build Run ID | Build # | Workflow ID | Progress | Status | Start Reason | Cancel Reason | Created | Started | Finished |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.BuildRunID),
		result.BuildNumber,
		escapeMarkdown(result.WorkflowID),
		escapeMarkdown(result.ExecutionProgress),
		escapeMarkdown(result.CompletionStatus),
		escapeMarkdown(result.StartReason),
		escapeMarkdown(result.CancelReason),
		escapeMarkdown(result.CreatedDate),
		escapeMarkdown(result.StartedDate),
		escapeMarkdown(result.FinishedDate),
	)
	return nil
}

func printCiProductsTable(resp *CiProductsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tBundle ID\tType\tCreated")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.Name,
			item.Attributes.BundleID,
			item.Attributes.ProductType,
			item.Attributes.CreatedDate,
		)
	}
	return w.Flush()
}

func printCiProductsMarkdown(resp *CiProductsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Bundle ID | Type | Created |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.BundleID),
			escapeMarkdown(item.Attributes.ProductType),
			escapeMarkdown(item.Attributes.CreatedDate),
		)
	}
	return nil
}

func printCiWorkflowsTable(resp *CiWorkflowsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tEnabled\tLast Modified")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%t\t%s\n",
			item.ID,
			item.Attributes.Name,
			item.Attributes.IsEnabled,
			item.Attributes.LastModifiedDate,
		)
	}
	return w.Flush()
}

func printCiWorkflowsMarkdown(resp *CiWorkflowsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Enabled | Last Modified |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %t | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Name),
			item.Attributes.IsEnabled,
			escapeMarkdown(item.Attributes.LastModifiedDate),
		)
	}
	return nil
}

func printCiBuildRunsTable(resp *CiBuildRunsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tBuild #\tProgress\tStatus\tStart Reason\tCreated\tStarted\tFinished")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.Number,
			string(item.Attributes.ExecutionProgress),
			string(item.Attributes.CompletionStatus),
			item.Attributes.StartReason,
			item.Attributes.CreatedDate,
			item.Attributes.StartedDate,
			item.Attributes.FinishedDate,
		)
	}
	return w.Flush()
}

func printCiBuildRunsMarkdown(resp *CiBuildRunsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Build # | Progress | Status | Start Reason | Created | Started | Finished |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %d | %s | %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			item.Attributes.Number,
			escapeMarkdown(string(item.Attributes.ExecutionProgress)),
			escapeMarkdown(string(item.Attributes.CompletionStatus)),
			escapeMarkdown(item.Attributes.StartReason),
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(item.Attributes.StartedDate),
			escapeMarkdown(item.Attributes.FinishedDate),
		)
	}
	return nil
}

func printCiBuildActionsTable(resp *CiBuildActionsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tType\tProgress\tStatus\tErrors\tWarnings\tStarted\tFinished")
	for _, item := range resp.Data {
		errors := 0
		warnings := 0
		if item.Attributes.IssueCounts != nil {
			errors = item.Attributes.IssueCounts.Errors
			warnings = item.Attributes.IssueCounts.Warnings
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
			item.Attributes.Name,
			item.Attributes.ActionType,
			string(item.Attributes.ExecutionProgress),
			string(item.Attributes.CompletionStatus),
			errors,
			warnings,
			item.Attributes.StartedDate,
			item.Attributes.FinishedDate,
		)
	}
	return w.Flush()
}

func printCiBuildActionsMarkdown(resp *CiBuildActionsResponse) error {
	fmt.Fprintln(os.Stdout, "| Name | Type | Progress | Status | Errors | Warnings | Started | Finished |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		errors := 0
		warnings := 0
		if item.Attributes.IssueCounts != nil {
			errors = item.Attributes.IssueCounts.Errors
			warnings = item.Attributes.IssueCounts.Warnings
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %d | %d | %s | %s |\n",
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.ActionType),
			escapeMarkdown(string(item.Attributes.ExecutionProgress)),
			escapeMarkdown(string(item.Attributes.CompletionStatus)),
			errors,
			warnings,
			escapeMarkdown(item.Attributes.StartedDate),
			escapeMarkdown(item.Attributes.FinishedDate),
		)
	}
	return nil
}

func printCiArtifactsTable(resp *CiArtifactsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tType\tSize\tDownload URL")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			item.ID,
			item.Attributes.FileName,
			item.Attributes.FileType,
			item.Attributes.FileSize,
			item.Attributes.DownloadURL,
		)
	}
	return w.Flush()
}

func printCiArtifactsMarkdown(resp *CiArtifactsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Type | Size | Download URL |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %d | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.FileName),
			escapeMarkdown(item.Attributes.FileType),
			item.Attributes.FileSize,
			escapeMarkdown(item.Attributes.DownloadURL),
		)
	}
	return nil
}

func printCiArtifactTable(resp *CiArtifactResponse) error {
	return printCiArtifactsTable(&CiArtifactsResponse{Data: []CiArtifactResource{resp.Data}})
}

func printCiArtifactMarkdown(resp *CiArtifactResponse) error {
	return printCiArtifactsMarkdown(&CiArtifactsResponse{Data: []CiArtifactResource{resp.Data}})
}

func printCiTestResultsTable(resp *CiTestResultsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tClass\tName\tStatus\tDuration")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.ClassName,
			item.Attributes.Name,
			string(item.Attributes.Status),
			formatTestDuration(item),
		)
	}
	return w.Flush()
}

func printCiTestResultsMarkdown(resp *CiTestResultsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Class | Name | Status | Duration |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.ClassName),
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(string(item.Attributes.Status)),
			escapeMarkdown(formatTestDuration(item)),
		)
	}
	return nil
}

func printCiTestResultTable(resp *CiTestResultResponse) error {
	return printCiTestResultsTable(&CiTestResultsResponse{Data: []CiTestResultResource{resp.Data}})
}

func printCiTestResultMarkdown(resp *CiTestResultResponse) error {
	return printCiTestResultsMarkdown(&CiTestResultsResponse{Data: []CiTestResultResource{resp.Data}})
}

func printCiIssuesTable(resp *CiIssuesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tFile\tLine\tMessage")
	for _, item := range resp.Data {
		filePath, lineNumber := formatFileLocation(item.Attributes.FileSource)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.IssueType,
			filePath,
			lineNumber,
			item.Attributes.Message,
		)
	}
	return w.Flush()
}

func printCiIssuesMarkdown(resp *CiIssuesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Type | File | Line | Message |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		filePath, lineNumber := formatFileLocation(item.Attributes.FileSource)
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.IssueType),
			escapeMarkdown(filePath),
			escapeMarkdown(lineNumber),
			escapeMarkdown(item.Attributes.Message),
		)
	}
	return nil
}

func printCiIssueTable(resp *CiIssueResponse) error {
	return printCiIssuesTable(&CiIssuesResponse{Data: []CiIssueResource{resp.Data}})
}

func printCiIssueMarkdown(resp *CiIssueResponse) error {
	return printCiIssuesMarkdown(&CiIssuesResponse{Data: []CiIssueResource{resp.Data}})
}

func printCiArtifactDownloadResultTable(result *CiArtifactDownloadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tType\tSize\tBytes Written\tOutput Path")
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
		result.ID,
		result.FileName,
		result.FileType,
		result.FileSize,
		result.BytesWritten,
		result.OutputPath,
	)
	return w.Flush()
}

func printCiArtifactDownloadResultMarkdown(result *CiArtifactDownloadResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Type | Size | Bytes Written | Output Path |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %d | %d | %s |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.FileName),
		escapeMarkdown(result.FileType),
		result.FileSize,
		result.BytesWritten,
		escapeMarkdown(result.OutputPath),
	)
	return nil
}

func formatTestDuration(result CiTestResultResource) string {
	if len(result.Attributes.DestinationTestResults) == 0 {
		return ""
	}
	duration := result.Attributes.DestinationTestResults[0].Duration
	if duration <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2fs", duration)
}

func formatFileLocation(location *FileLocation) (string, string) {
	if location == nil {
		return "", ""
	}
	line := ""
	if location.LineNumber > 0 {
		line = fmt.Sprintf("%d", location.LineNumber)
	}
	return location.Path, line
}
