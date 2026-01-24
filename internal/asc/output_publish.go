package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func printTestFlightPublishResultTable(result *TestFlightPublishResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Build ID\tVersion\tBuild Number\tProcessing\tGroups\tUploaded\tNotified")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
		result.BuildID,
		result.BuildVersion,
		result.BuildNumber,
		result.ProcessingState,
		strings.Join(result.GroupIDs, ", "),
		result.Uploaded,
		result.Notified,
	)
	return w.Flush()
}

func printTestFlightPublishResultMarkdown(result *TestFlightPublishResult) error {
	fmt.Fprintln(os.Stdout, "| Build ID | Version | Build Number | Processing | Groups | Uploaded | Notified |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %t | %t |\n",
		escapeMarkdown(result.BuildID),
		escapeMarkdown(result.BuildVersion),
		escapeMarkdown(result.BuildNumber),
		escapeMarkdown(result.ProcessingState),
		escapeMarkdown(strings.Join(result.GroupIDs, ", ")),
		result.Uploaded,
		result.Notified,
	)
	return nil
}

func printAppStorePublishResultTable(result *AppStorePublishResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Build ID\tVersion ID\tSubmission ID\tUploaded\tAttached\tSubmitted")
	fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%t\t%t\n",
		result.BuildID,
		result.VersionID,
		result.SubmissionID,
		result.Uploaded,
		result.Attached,
		result.Submitted,
	)
	return w.Flush()
}

func printAppStorePublishResultMarkdown(result *AppStorePublishResult) error {
	fmt.Fprintln(os.Stdout, "| Build ID | Version ID | Submission ID | Uploaded | Attached | Submitted |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %t | %t | %t |\n",
		escapeMarkdown(result.BuildID),
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.SubmissionID),
		result.Uploaded,
		result.Attached,
		result.Submitted,
	)
	return nil
}
