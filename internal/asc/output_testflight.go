package asc

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

func formatBetaReviewContactName(attr BetaAppReviewDetailAttributes) string {
	first := strings.TrimSpace(attr.ContactFirstName)
	last := strings.TrimSpace(attr.ContactLastName)
	switch {
	case first == "" && last == "":
		return ""
	case first == "":
		return last
	case last == "":
		return first
	default:
		return first + " " + last
	}
}

func printBetaAppReviewDetailsTable(resp *BetaAppReviewDetailsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tContact\tEmail\tPhone\tDemo Required\tDemo Account\tNotes")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
			item.ID,
			compactWhitespace(formatBetaReviewContactName(item.Attributes)),
			compactWhitespace(item.Attributes.ContactEmail),
			compactWhitespace(item.Attributes.ContactPhone),
			item.Attributes.DemoAccountRequired,
			compactWhitespace(item.Attributes.DemoAccountName),
			compactWhitespace(item.Attributes.Notes),
		)
	}
	return w.Flush()
}

func printBetaAppReviewDetailTable(resp *BetaAppReviewDetailResponse) error {
	return printBetaAppReviewDetailsTable(&BetaAppReviewDetailsResponse{
		Data: []Resource[BetaAppReviewDetailAttributes]{resp.Data},
	})
}

func printBetaAppReviewDetailsMarkdown(resp *BetaAppReviewDetailsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Contact | Email | Phone | Demo Required | Demo Account | Notes |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %t | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(formatBetaReviewContactName(item.Attributes)),
			escapeMarkdown(item.Attributes.ContactEmail),
			escapeMarkdown(item.Attributes.ContactPhone),
			item.Attributes.DemoAccountRequired,
			escapeMarkdown(item.Attributes.DemoAccountName),
			escapeMarkdown(item.Attributes.Notes),
		)
	}
	return nil
}

func printBetaAppReviewDetailMarkdown(resp *BetaAppReviewDetailResponse) error {
	return printBetaAppReviewDetailsMarkdown(&BetaAppReviewDetailsResponse{
		Data: []Resource[BetaAppReviewDetailAttributes]{resp.Data},
	})
}

func printBetaAppReviewSubmissionsTable(resp *BetaAppReviewSubmissionsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tState\tSubmitted Date")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.BetaReviewState),
			compactWhitespace(item.Attributes.SubmittedDate),
		)
	}
	return w.Flush()
}

func printBetaAppReviewSubmissionTable(resp *BetaAppReviewSubmissionResponse) error {
	return printBetaAppReviewSubmissionsTable(&BetaAppReviewSubmissionsResponse{
		Data: []Resource[BetaAppReviewSubmissionAttributes]{resp.Data},
	})
}

func printBetaAppReviewSubmissionsMarkdown(resp *BetaAppReviewSubmissionsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | State | Submitted Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.BetaReviewState),
			escapeMarkdown(item.Attributes.SubmittedDate),
		)
	}
	return nil
}

func printBetaAppReviewSubmissionMarkdown(resp *BetaAppReviewSubmissionResponse) error {
	return printBetaAppReviewSubmissionsMarkdown(&BetaAppReviewSubmissionsResponse{
		Data: []Resource[BetaAppReviewSubmissionAttributes]{resp.Data},
	})
}

func printBuildBetaDetailsTable(resp *BuildBetaDetailsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAuto Notify\tInternal State\tExternal State")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%t\t%s\t%s\n",
			item.ID,
			item.Attributes.AutoNotifyEnabled,
			compactWhitespace(item.Attributes.InternalBuildState),
			compactWhitespace(item.Attributes.ExternalBuildState),
		)
	}
	return w.Flush()
}

func printBuildBetaDetailTable(resp *BuildBetaDetailResponse) error {
	return printBuildBetaDetailsTable(&BuildBetaDetailsResponse{
		Data: []Resource[BuildBetaDetailAttributes]{resp.Data},
	})
}

func printBuildBetaDetailsMarkdown(resp *BuildBetaDetailsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Auto Notify | Internal State | External State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %t | %s | %s |\n",
			escapeMarkdown(item.ID),
			item.Attributes.AutoNotifyEnabled,
			escapeMarkdown(item.Attributes.InternalBuildState),
			escapeMarkdown(item.Attributes.ExternalBuildState),
		)
	}
	return nil
}

func printBuildBetaDetailMarkdown(resp *BuildBetaDetailResponse) error {
	return printBuildBetaDetailsMarkdown(&BuildBetaDetailsResponse{
		Data: []Resource[BuildBetaDetailAttributes]{resp.Data},
	})
}

func printBetaRecruitmentCriterionOptionsTable(resp *BetaRecruitmentCriterionOptionsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tIdentifier\tName\tCategory")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Identifier),
			compactWhitespace(item.Attributes.Name),
			compactWhitespace(item.Attributes.Category),
		)
	}
	return w.Flush()
}

func printBetaRecruitmentCriterionOptionsMarkdown(resp *BetaRecruitmentCriterionOptionsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Identifier | Name | Category |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Identifier),
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.Category),
		)
	}
	return nil
}

func printBetaRecruitmentCriteriaTable(resp *BetaRecruitmentCriteriaResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLast Modified")
	fmt.Fprintf(w, "%s\t%s\n",
		resp.Data.ID,
		compactWhitespace(resp.Data.Attributes.LastModifiedDate),
	)
	return w.Flush()
}

func printBetaRecruitmentCriteriaMarkdown(resp *BetaRecruitmentCriteriaResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Last Modified |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s |\n",
		escapeMarkdown(resp.Data.ID),
		escapeMarkdown(resp.Data.Attributes.LastModifiedDate),
	)
	return nil
}

func formatMetricAttributes(attrs BetaGroupMetricAttributes) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, attrs[key]))
	}
	return strings.Join(parts, ", ")
}

func printBetaGroupMetricsTable(items []Resource[BetaGroupMetricAttributes]) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAttributes")
	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\n",
			item.ID,
			compactWhitespace(formatMetricAttributes(item.Attributes)),
		)
	}
	return w.Flush()
}

func printBetaGroupMetricsMarkdown(items []Resource[BetaGroupMetricAttributes]) error {
	fmt.Fprintln(os.Stdout, "| ID | Attributes |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range items {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(formatMetricAttributes(item.Attributes)),
		)
	}
	return nil
}
