package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// BuildUploadResult represents CLI output for build upload operations.
type BuildUploadResult struct {
	UploadID            string            `json:"uploadId"`
	FileID              string            `json:"fileId"`
	FileName            string            `json:"fileName"`
	FileSize            int64             `json:"fileSize"`
	Operations          []UploadOperation `json:"operations,omitempty"`
	Uploaded            *bool             `json:"uploaded,omitempty"`
	ChecksumVerified    *bool             `json:"checksumVerified,omitempty"`
	SourceFileChecksums *Checksums        `json:"sourceFileChecksums,omitempty"`
}

// BuildBetaGroupsUpdateResult represents CLI output for build beta group updates.
type BuildBetaGroupsUpdateResult struct {
	BuildID  string   `json:"buildId"`
	GroupIDs []string `json:"groupIds"`
	Action   string   `json:"action"`
}

func printBuildsTable(resp *BuildsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tUploaded\tProcessing\tExpired")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\n",
			item.Attributes.Version,
			item.Attributes.UploadedDate,
			item.Attributes.ProcessingState,
			item.Attributes.Expired,
		)
	}
	return w.Flush()
}

func printBuildsMarkdown(resp *BuildsResponse) error {
	fmt.Fprintln(os.Stdout, "| Version | Uploaded | Processing | Expired |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %t |\n",
			escapeMarkdown(item.Attributes.Version),
			escapeMarkdown(item.Attributes.UploadedDate),
			escapeMarkdown(item.Attributes.ProcessingState),
			item.Attributes.Expired,
		)
	}
	return nil
}

func printBuildUploadResultTable(result *BuildUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	headers := []string{"Upload ID", "File ID", "File Name", "File Size"}
	values := []string{
		result.UploadID,
		result.FileID,
		result.FileName,
		fmt.Sprintf("%d", result.FileSize),
	}
	if result.Uploaded != nil {
		headers = append(headers, "Uploaded")
		values = append(values, fmt.Sprintf("%t", *result.Uploaded))
	}
	if result.ChecksumVerified != nil {
		headers = append(headers, "Checksum Verified")
		values = append(values, fmt.Sprintf("%t", *result.ChecksumVerified))
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Join(values, "\t"))
	if err := w.Flush(); err != nil {
		return err
	}
	if len(result.Operations) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\nUpload Operations")
	opsWriter := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(opsWriter, "Method\tURL\tLength\tOffset")
	for _, op := range result.Operations {
		fmt.Fprintf(opsWriter, "%s\t%s\t%d\t%d\n",
			op.Method,
			op.URL,
			op.Length,
			op.Offset,
		)
	}
	return opsWriter.Flush()
}

func printBuildUploadResultMarkdown(result *BuildUploadResult) error {
	headers := []string{"Upload ID", "File ID", "File Name", "File Size"}
	values := []string{
		escapeMarkdown(result.UploadID),
		escapeMarkdown(result.FileID),
		escapeMarkdown(result.FileName),
		fmt.Sprintf("%d", result.FileSize),
	}
	if result.Uploaded != nil {
		headers = append(headers, "Uploaded")
		values = append(values, fmt.Sprintf("%t", *result.Uploaded))
	}
	if result.ChecksumVerified != nil {
		headers = append(headers, "Checksum Verified")
		values = append(values, fmt.Sprintf("%t", *result.ChecksumVerified))
	}
	separator := make([]string, len(headers))
	for i := range separator {
		separator[i] = "---"
	}
	fmt.Fprintf(os.Stdout, "| %s |\n", strings.Join(headers, " | "))
	fmt.Fprintf(os.Stdout, "| %s |\n", strings.Join(separator, " | "))
	fmt.Fprintf(os.Stdout, "| %s |\n", strings.Join(values, " | "))
	if len(result.Operations) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\n| Method | URL | Length | Offset |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, op := range result.Operations {
		fmt.Fprintf(os.Stdout, "| %s | %s | %d | %d |\n",
			escapeMarkdown(op.Method),
			escapeMarkdown(op.URL),
			op.Length,
			op.Offset,
		)
	}
	return nil
}

func printBuildBetaGroupsUpdateTable(result *BuildBetaGroupsUpdateResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Build ID\tGroup IDs\tAction")
	fmt.Fprintf(w, "%s\t%s\t%s\n",
		result.BuildID,
		strings.Join(result.GroupIDs, ", "),
		result.Action,
	)
	return w.Flush()
}

func printBuildBetaGroupsUpdateMarkdown(result *BuildBetaGroupsUpdateResult) error {
	fmt.Fprintln(os.Stdout, "| Build ID | Group IDs | Action |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
		escapeMarkdown(result.BuildID),
		escapeMarkdown(strings.Join(result.GroupIDs, ", ")),
		escapeMarkdown(result.Action),
	)
	return nil
}
