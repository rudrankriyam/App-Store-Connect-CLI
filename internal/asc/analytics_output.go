package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// SalesReportResult represents CLI output for sales report downloads.
type SalesReportResult struct {
	VendorNumber     string `json:"vendorNumber"`
	ReportType       string `json:"reportType"`
	ReportSubType    string `json:"reportSubType"`
	Frequency        string `json:"frequency"`
	ReportDate       string `json:"reportDate"`
	Version          string `json:"version,omitempty"`
	FilePath         string `json:"filePath"`
	FileSize         int64  `json:"fileSize"`
	Decompressed     bool   `json:"decompressed"`
	DecompressedPath string `json:"decompressedPath,omitempty"`
	DecompressedSize int64  `json:"decompressedSize,omitempty"`
}

// AnalyticsReportRequestResult represents CLI output for created requests.
type AnalyticsReportRequestResult struct {
	RequestID   string `json:"requestId"`
	AppID       string `json:"appId"`
	AccessType  string `json:"accessType"`
	State       string `json:"state,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
}

// AnalyticsReportDownloadResult represents CLI output for analytics downloads.
type AnalyticsReportDownloadResult struct {
	RequestID        string `json:"requestId"`
	InstanceID       string `json:"instanceId"`
	SegmentID        string `json:"segmentId,omitempty"`
	FilePath         string `json:"filePath"`
	FileSize         int64  `json:"fileSize"`
	Decompressed     bool   `json:"decompressed"`
	DecompressedPath string `json:"decompressedPath,omitempty"`
	DecompressedSize int64  `json:"decompressedSize,omitempty"`
}

// AnalyticsReportGetResult represents CLI output for report metadata with instances.
type AnalyticsReportGetResult struct {
	RequestID string                     `json:"requestId"`
	Data      []AnalyticsReportGetReport `json:"data"`
	Links     Links                      `json:"links,omitempty"`
}

// AnalyticsReportGetReport represents an analytics report with instances.
type AnalyticsReportGetReport struct {
	ID          string                       `json:"id"`
	ReportType  string                       `json:"reportType,omitempty"`
	Name        string                       `json:"name,omitempty"`
	Category    string                       `json:"category,omitempty"`
	Granularity string                       `json:"granularity,omitempty"`
	Instances   []AnalyticsReportGetInstance `json:"instances,omitempty"`
}

// AnalyticsReportGetInstance represents a report instance with segments.
type AnalyticsReportGetInstance struct {
	ID             string                      `json:"id"`
	ReportDate     string                      `json:"reportDate,omitempty"`
	ProcessingDate string                      `json:"processingDate,omitempty"`
	Granularity    string                      `json:"granularity,omitempty"`
	Version        string                      `json:"version,omitempty"`
	Segments       []AnalyticsReportGetSegment `json:"segments,omitempty"`
}

// AnalyticsReportGetSegment represents a report segment with download URL.
type AnalyticsReportGetSegment struct {
	ID                string `json:"id"`
	DownloadURL       string `json:"downloadUrl,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
	SizeInBytes       int64  `json:"sizeInBytes,omitempty"`
	URLExpirationDate string `json:"urlExpirationDate,omitempty"`
}

func printSalesReportResultTable(result *SalesReportResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Vendor\tType\tSubtype\tFrequency\tDate\tVersion\tCompressed File\tCompressed Size\tDecompressed File\tDecompressed Size")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\n",
		result.VendorNumber,
		result.ReportType,
		result.ReportSubType,
		result.Frequency,
		result.ReportDate,
		result.Version,
		result.FilePath,
		result.FileSize,
		result.DecompressedPath,
		result.DecompressedSize,
	)
	return w.Flush()
}

func printSalesReportResultMarkdown(result *SalesReportResult) error {
	fmt.Fprintln(os.Stdout, "| Vendor | Type | Subtype | Frequency | Date | Version | Compressed File | Compressed Size | Decompressed File | Decompressed Size |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s | %s | %d | %s | %d |\n",
		escapeMarkdown(result.VendorNumber),
		escapeMarkdown(result.ReportType),
		escapeMarkdown(result.ReportSubType),
		escapeMarkdown(result.Frequency),
		escapeMarkdown(result.ReportDate),
		escapeMarkdown(result.Version),
		escapeMarkdown(result.FilePath),
		result.FileSize,
		escapeMarkdown(result.DecompressedPath),
		result.DecompressedSize,
	)
	return nil
}

func printAnalyticsReportRequestResultTable(result *AnalyticsReportRequestResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Request ID\tApp ID\tAccess Type\tState\tCreated Date")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		result.RequestID,
		result.AppID,
		result.AccessType,
		result.State,
		result.CreatedDate,
	)
	return w.Flush()
}

func printAnalyticsReportRequestResultMarkdown(result *AnalyticsReportRequestResult) error {
	fmt.Fprintln(os.Stdout, "| Request ID | App ID | Access Type | State | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.RequestID),
		escapeMarkdown(result.AppID),
		escapeMarkdown(result.AccessType),
		escapeMarkdown(result.State),
		escapeMarkdown(result.CreatedDate),
	)
	return nil
}

func printAnalyticsReportRequestsTable(resp *AnalyticsReportRequestsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAccess Type\tState\tCreated Date\tApp ID")
	for _, item := range resp.Data {
		appID := ""
		if item.Relationships != nil && item.Relationships.App != nil {
			appID = item.Relationships.App.Data.ID
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			string(item.Attributes.AccessType),
			string(item.Attributes.State),
			item.Attributes.CreatedDate,
			appID,
		)
	}
	return w.Flush()
}

func printAnalyticsReportRequestsMarkdown(resp *AnalyticsReportRequestsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Access Type | State | Created Date | App ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		appID := ""
		if item.Relationships != nil && item.Relationships.App != nil {
			appID = item.Relationships.App.Data.ID
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(string(item.Attributes.AccessType)),
			escapeMarkdown(string(item.Attributes.State)),
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(appID),
		)
	}
	return nil
}

func printAnalyticsReportDownloadResultTable(result *AnalyticsReportDownloadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Request ID\tInstance ID\tSegment ID\tCompressed File\tCompressed Size\tDecompressed File\tDecompressed Size")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%d\n",
		result.RequestID,
		result.InstanceID,
		result.SegmentID,
		result.FilePath,
		result.FileSize,
		result.DecompressedPath,
		result.DecompressedSize,
	)
	return w.Flush()
}

func printAnalyticsReportDownloadResultMarkdown(result *AnalyticsReportDownloadResult) error {
	fmt.Fprintln(os.Stdout, "| Request ID | Instance ID | Segment ID | Compressed File | Compressed Size | Decompressed File | Decompressed Size |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %d | %s | %d |\n",
		escapeMarkdown(result.RequestID),
		escapeMarkdown(result.InstanceID),
		escapeMarkdown(result.SegmentID),
		escapeMarkdown(result.FilePath),
		result.FileSize,
		escapeMarkdown(result.DecompressedPath),
		result.DecompressedSize,
	)
	return nil
}

func printAnalyticsReportGetResultTable(result *AnalyticsReportGetResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Report ID\tName\tCategory\tGranularity\tInstances\tSegments")
	for _, report := range result.Data {
		name := report.Name
		if name == "" {
			name = report.ReportType
		}
		segments := countSegments(report.Instances)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\n",
			report.ID,
			name,
			report.Category,
			report.Granularity,
			len(report.Instances),
			segments,
		)
	}
	return w.Flush()
}

func printAnalyticsReportGetResultMarkdown(result *AnalyticsReportGetResult) error {
	fmt.Fprintln(os.Stdout, "| Report ID | Name | Category | Granularity | Instances | Segments |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	for _, report := range result.Data {
		name := report.Name
		if name == "" {
			name = report.ReportType
		}
		segments := countSegments(report.Instances)
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %d | %d |\n",
			escapeMarkdown(report.ID),
			escapeMarkdown(name),
			escapeMarkdown(report.Category),
			escapeMarkdown(report.Granularity),
			len(report.Instances),
			segments,
		)
	}
	return nil
}

func countSegments(instances []AnalyticsReportGetInstance) int {
	total := 0
	for _, instance := range instances {
		total += len(instance.Segments)
	}
	return total
}
