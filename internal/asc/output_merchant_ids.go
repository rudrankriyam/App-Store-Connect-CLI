package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// MerchantIDDeleteResult represents CLI output for merchant ID deletions.
type MerchantIDDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

func printMerchantIDsTable(resp *MerchantIDsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tIdentifier")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.Identifier,
		)
	}
	return w.Flush()
}

func printMerchantIDsMarkdown(resp *MerchantIDsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Identifier |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.Identifier),
		)
	}
	return nil
}

func printMerchantIDDeleteResultTable(result *MerchantIDDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printMerchantIDDeleteResultMarkdown(result *MerchantIDDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}
