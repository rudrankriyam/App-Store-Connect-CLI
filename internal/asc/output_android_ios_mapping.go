package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// AndroidToIosAppMappingDeleteResult represents CLI output for deletions.
type AndroidToIosAppMappingDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

func printAndroidToIosAppMappingDetailsTable(resp *AndroidToIosAppMappingDetailsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPackage Name\tFingerprints")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			item.Attributes.PackageName,
			formatAndroidToIosFingerprints(item.Attributes.AppSigningKeyPublicCertificateSha256Fingerprints),
		)
	}
	return w.Flush()
}

func printAndroidToIosAppMappingDetailsMarkdown(resp *AndroidToIosAppMappingDetailsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Package Name | Fingerprints |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.PackageName),
			escapeMarkdown(formatAndroidToIosFingerprints(item.Attributes.AppSigningKeyPublicCertificateSha256Fingerprints)),
		)
	}
	return nil
}

func printAndroidToIosAppMappingDeleteResultTable(result *AndroidToIosAppMappingDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAndroidToIosAppMappingDeleteResultMarkdown(result *AndroidToIosAppMappingDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func formatAndroidToIosFingerprints(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}
