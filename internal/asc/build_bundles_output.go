package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolValue(value *bool) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%t", *value)
}

func int64Value(value *int64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

func buildBundleTypeValue(value *BuildBundleType) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

func printBuildBundlesTable(resp *BuildBundlesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tBundle ID\tType\tFile Name\tSDK Build\tPlatform Build")
	for _, item := range resp.Data {
		attrs := item.Attributes
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			stringValue(attrs.BundleID),
			buildBundleTypeValue(attrs.BundleType),
			stringValue(attrs.FileName),
			stringValue(attrs.SDKBuild),
			stringValue(attrs.PlatformBuild),
		)
	}
	return w.Flush()
}

func printBuildBundlesMarkdown(resp *BuildBundlesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Bundle ID | Type | File Name | SDK Build | Platform Build |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		attrs := item.Attributes
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(stringValue(attrs.BundleID)),
			escapeMarkdown(buildBundleTypeValue(attrs.BundleType)),
			escapeMarkdown(stringValue(attrs.FileName)),
			escapeMarkdown(stringValue(attrs.SDKBuild)),
			escapeMarkdown(stringValue(attrs.PlatformBuild)),
		)
	}
	return nil
}

func printBuildBundleFileSizesTable(resp *BuildBundleFileSizesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDevice Model\tOS Version\tDownload Bytes\tInstall Bytes")
	for _, item := range resp.Data {
		attrs := item.Attributes
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			stringValue(attrs.DeviceModel),
			stringValue(attrs.OSVersion),
			int64Value(attrs.DownloadBytes),
			int64Value(attrs.InstallBytes),
		)
	}
	return w.Flush()
}

func printBuildBundleFileSizesMarkdown(resp *BuildBundleFileSizesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Device Model | OS Version | Download Bytes | Install Bytes |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		attrs := item.Attributes
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(stringValue(attrs.DeviceModel)),
			escapeMarkdown(stringValue(attrs.OSVersion)),
			escapeMarkdown(int64Value(attrs.DownloadBytes)),
			escapeMarkdown(int64Value(attrs.InstallBytes)),
		)
	}
	return nil
}

func printBetaAppClipInvocationsTable(resp *BetaAppClipInvocationsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tURL")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", item.ID, stringValue(item.Attributes.URL))
	}
	return w.Flush()
}

func printBetaAppClipInvocationsMarkdown(resp *BetaAppClipInvocationsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | URL |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(stringValue(item.Attributes.URL)),
		)
	}
	return nil
}

func printAppClipDomainStatusResultTable(result *AppClipDomainStatusResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Build Bundle ID\tAvailable\tStatus ID\tLast Updated")
	fmt.Fprintf(w, "%s\t%t\t%s\t%s\n",
		result.BuildBundleID,
		result.Available,
		result.StatusID,
		stringValue(result.LastUpdatedDate),
	)
	if err := w.Flush(); err != nil {
		return err
	}
	if len(result.Domains) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\nDomains")
	domains := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(domains, "Domain\tValid\tLast Updated\tError")
	for _, domain := range result.Domains {
		fmt.Fprintf(domains, "%s\t%s\t%s\t%s\n",
			stringValue(domain.Domain),
			boolValue(domain.IsValid),
			stringValue(domain.LastUpdatedDate),
			stringValue(domain.ErrorCode),
		)
	}
	return domains.Flush()
}

func printAppClipDomainStatusResultMarkdown(result *AppClipDomainStatusResult) error {
	fmt.Fprintln(os.Stdout, "| Build Bundle ID | Available | Status ID | Last Updated |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t | %s | %s |\n",
		escapeMarkdown(result.BuildBundleID),
		result.Available,
		escapeMarkdown(result.StatusID),
		escapeMarkdown(stringValue(result.LastUpdatedDate)),
	)
	if len(result.Domains) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\n| Domain | Valid | Last Updated | Error |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, domain := range result.Domains {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
			escapeMarkdown(stringValue(domain.Domain)),
			boolValue(domain.IsValid),
			escapeMarkdown(stringValue(domain.LastUpdatedDate)),
			escapeMarkdown(stringValue(domain.ErrorCode)),
		)
	}
	return nil
}
