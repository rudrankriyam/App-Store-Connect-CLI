package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type routingAppCoverageField struct {
	Name  string
	Value string
}

func printRoutingAppCoverageTable(resp *RoutingAppCoverageResponse) error {
	fields := routingAppCoverageFields(resp)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Field\tValue")
	for _, field := range fields {
		fmt.Fprintf(w, "%s\t%s\n", field.Name, field.Value)
	}
	return w.Flush()
}

func printRoutingAppCoverageMarkdown(resp *RoutingAppCoverageResponse) error {
	fields := routingAppCoverageFields(resp)
	fmt.Fprintln(os.Stdout, "| Field | Value |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, field := range fields {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n", escapeMarkdown(field.Name), escapeMarkdown(field.Value))
	}
	return nil
}

func routingAppCoverageFields(resp *RoutingAppCoverageResponse) []routingAppCoverageField {
	if resp == nil {
		return nil
	}
	attrs := resp.Data.Attributes
	return []routingAppCoverageField{
		{Name: "ID", Value: fallbackValue(resp.Data.ID)},
		{Name: "Type", Value: fallbackValue(string(resp.Data.Type))},
		{Name: "File Name", Value: fallbackValue(attrs.FileName)},
		{Name: "File Size", Value: formatAttachmentFileSize(attrs.FileSize)},
		{Name: "Source File Checksum", Value: fallbackValue(attrs.SourceFileChecksum)},
		{Name: "Delivery State", Value: formatAssetDeliveryState(attrs.AssetDeliveryState)},
	}
}

func printRoutingAppCoverageDeleteResultTable(result *RoutingAppCoverageDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printRoutingAppCoverageDeleteResultMarkdown(result *RoutingAppCoverageDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n", escapeMarkdown(result.ID), result.Deleted)
	return nil
}
