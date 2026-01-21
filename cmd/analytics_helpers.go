package cmd

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

const analyticsMaxLimit = 200

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func resolveVendorNumber(value string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	if env := strings.TrimSpace(os.Getenv("ASC_VENDOR_NUMBER")); env != "" {
		return env
	}
	return strings.TrimSpace(os.Getenv("ASC_ANALYTICS_VENDOR_NUMBER"))
}

func normalizeSalesReportType(value string) (asc.SalesReportType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(asc.SalesReportTypeSales):
		return asc.SalesReportTypeSales, nil
	case string(asc.SalesReportTypePreOrder):
		return asc.SalesReportTypePreOrder, nil
	case string(asc.SalesReportTypeNewsstand):
		return asc.SalesReportTypeNewsstand, nil
	case string(asc.SalesReportTypeSubscription):
		return asc.SalesReportTypeSubscription, nil
	case string(asc.SalesReportTypeSubscriptionEvent):
		return asc.SalesReportTypeSubscriptionEvent, nil
	default:
		return "", fmt.Errorf("--type must be SALES, PRE_ORDER, NEWSSTAND, SUBSCRIPTION, or SUBSCRIPTION_EVENT")
	}
}

func normalizeSalesReportSubType(value string) (asc.SalesReportSubType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(asc.SalesReportSubTypeSummary):
		return asc.SalesReportSubTypeSummary, nil
	case string(asc.SalesReportSubTypeDetailed):
		return asc.SalesReportSubTypeDetailed, nil
	default:
		return "", fmt.Errorf("--subtype must be SUMMARY or DETAILED")
	}
}

func normalizeSalesReportFrequency(value string) (asc.SalesReportFrequency, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(asc.SalesReportFrequencyDaily):
		return asc.SalesReportFrequencyDaily, nil
	case string(asc.SalesReportFrequencyWeekly):
		return asc.SalesReportFrequencyWeekly, nil
	case string(asc.SalesReportFrequencyMonthly):
		return asc.SalesReportFrequencyMonthly, nil
	case string(asc.SalesReportFrequencyYearly):
		return asc.SalesReportFrequencyYearly, nil
	default:
		return "", fmt.Errorf("--frequency must be DAILY, WEEKLY, MONTHLY, or YEARLY")
	}
}

func normalizeSalesReportVersion(value string) (asc.SalesReportVersion, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return asc.SalesReportVersion1_0, nil
	}
	switch normalized {
	case string(asc.SalesReportVersion1_0):
		return asc.SalesReportVersion1_0, nil
	case string(asc.SalesReportVersion1_1):
		return asc.SalesReportVersion1_1, nil
	default:
		return "", fmt.Errorf("--version must be 1_0 or 1_1")
	}
}

func normalizeAnalyticsAccessType(value string) (asc.AnalyticsAccessType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(asc.AnalyticsAccessTypeOngoing):
		return asc.AnalyticsAccessTypeOngoing, nil
	case string(asc.AnalyticsAccessTypeOneTimeSnapshot):
		return asc.AnalyticsAccessTypeOneTimeSnapshot, nil
	default:
		return "", fmt.Errorf("--access-type must be ONGOING or ONE_TIME_SNAPSHOT")
	}
}

func normalizeAnalyticsRequestState(value string) (asc.AnalyticsReportRequestState, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case string(asc.AnalyticsReportRequestStateProcessing):
		return asc.AnalyticsReportRequestStateProcessing, nil
	case string(asc.AnalyticsReportRequestStateCompleted):
		return asc.AnalyticsReportRequestStateCompleted, nil
	case string(asc.AnalyticsReportRequestStateFailed):
		return asc.AnalyticsReportRequestStateFailed, nil
	default:
		return "", fmt.Errorf("--state must be PROCESSING, COMPLETED, or FAILED")
	}
}

func validateUUIDFlag(flagName, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", flagName)
	}
	if !uuidPattern.MatchString(strings.TrimSpace(value)) {
		return fmt.Errorf("%s must be a valid UUID", flagName)
	}
	return nil
}

func normalizeReportDate(value string, frequency asc.SalesReportFrequency) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("--date is required")
	}
	switch frequency {
	case asc.SalesReportFrequencyMonthly:
		parsed, err := time.Parse("2006-01", trimmed)
		if err != nil {
			return "", fmt.Errorf("--date must be in YYYY-MM format for monthly reports")
		}
		return parsed.Format("2006-01"), nil
	case asc.SalesReportFrequencyYearly:
		parsed, err := time.Parse("2006", trimmed)
		if err != nil {
			return "", fmt.Errorf("--date must be in YYYY format for yearly reports")
		}
		return parsed.Format("2006"), nil
	default:
		parsed, err := time.Parse("2006-01-02", trimmed)
		if err != nil {
			return "", fmt.Errorf("--date must be in YYYY-MM-DD format")
		}
		return parsed.Format("2006-01-02"), nil
	}
}

func normalizeAnalyticsDateFilter(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return "", fmt.Errorf("--date must be in YYYY-MM-DD format")
	}
	return parsed.Format("2006-01-02"), nil
}

func resolveReportOutputPaths(outputPath, defaultCompressed, decompressedExt string, decompress bool) (string, string) {
	compressed := strings.TrimSpace(outputPath)
	if compressed == "" {
		compressed = defaultCompressed
	}
	if !decompress {
		return compressed, ""
	}
	if strings.HasSuffix(compressed, ".gz") {
		return compressed, strings.TrimSuffix(compressed, ".gz")
	}
	if strings.HasSuffix(compressed, decompressedExt) {
		return compressed + ".gz", compressed
	}
	return compressed, compressed + decompressedExt
}

func writeStreamToFile(path string, reader io.Reader) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	file, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return io.Copy(file, reader)
}

func decompressGzipFile(sourcePath, destPath string) (int64, error) {
	in, err := os.Open(sourcePath)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	reader, err := gzip.NewReader(in)
	if err != nil {
		return 0, fmt.Errorf("failed to decompress gzip: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return 0, err
	}
	out, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	return io.Copy(out, reader)
}

func matchAnalyticsInstanceDate(attrs asc.AnalyticsReportInstanceAttributes, date string) bool {
	if strings.TrimSpace(date) == "" {
		return true
	}
	if strings.HasPrefix(attrs.ReportDate, date) {
		return true
	}
	return strings.HasPrefix(attrs.ProcessingDate, date)
}

func fetchAnalyticsReports(ctx context.Context, client *asc.Client, requestID string, limit int, next string, paginate bool) ([]asc.Resource[asc.AnalyticsReportAttributes], asc.Links, error) {
	var (
		all   []asc.Resource[asc.AnalyticsReportAttributes]
		links asc.Links
		seen  = make(map[string]bool)
	)

	if strings.TrimSpace(next) != "" {
		resp, err := client.GetAnalyticsReports(ctx, requestID, asc.WithAnalyticsReportsNextURL(next))
		if err != nil {
			return nil, asc.Links{}, err
		}
		return resp.Data, resp.Links, nil
	}

	if limit <= 0 {
		limit = analyticsMaxLimit
	}
	nextURL := ""
	for {
		var resp *asc.AnalyticsReportsResponse
		var err error
		if nextURL != "" {
			if seen[nextURL] {
				return nil, asc.Links{}, fmt.Errorf("analytics get: detected repeated pagination URL")
			}
			seen[nextURL] = true
			resp, err = client.GetAnalyticsReports(ctx, requestID, asc.WithAnalyticsReportsNextURL(nextURL))
		} else {
			resp, err = client.GetAnalyticsReports(ctx, requestID, asc.WithAnalyticsReportsLimit(limit))
		}
		if err != nil {
			return nil, asc.Links{}, err
		}
		if links.Self == "" {
			links.Self = resp.Links.Self
		}
		all = append(all, resp.Data...)
		links.Next = resp.Links.Next
		if !paginate || resp.Links.Next == "" {
			break
		}
		nextURL = resp.Links.Next
	}
	return all, links, nil
}

func fetchAnalyticsReportInstances(ctx context.Context, client *asc.Client, reportID string) ([]asc.Resource[asc.AnalyticsReportInstanceAttributes], error) {
	var (
		all  []asc.Resource[asc.AnalyticsReportInstanceAttributes]
		next string
		seen = make(map[string]bool)
	)
	for {
		var resp *asc.AnalyticsReportInstancesResponse
		var err error
		if next != "" {
			if seen[next] {
				return nil, fmt.Errorf("analytics get: detected repeated instance pagination URL")
			}
			seen[next] = true
			resp, err = client.GetAnalyticsReportInstances(ctx, reportID, asc.WithAnalyticsReportInstancesNextURL(next))
		} else {
			resp, err = client.GetAnalyticsReportInstances(ctx, reportID, asc.WithAnalyticsReportInstancesLimit(analyticsMaxLimit))
		}
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if resp.Links.Next == "" {
			break
		}
		next = resp.Links.Next
	}
	return all, nil
}

func fetchAnalyticsReportSegments(ctx context.Context, client *asc.Client, instanceID string) ([]asc.Resource[asc.AnalyticsReportSegmentAttributes], error) {
	var (
		all  []asc.Resource[asc.AnalyticsReportSegmentAttributes]
		next string
		seen = make(map[string]bool)
	)
	for {
		var resp *asc.AnalyticsReportSegmentsResponse
		var err error
		if next != "" {
			if seen[next] {
				return nil, fmt.Errorf("analytics get: detected repeated segment pagination URL")
			}
			seen[next] = true
			resp, err = client.GetAnalyticsReportSegments(ctx, instanceID, asc.WithAnalyticsReportSegmentsNextURL(next))
		} else {
			resp, err = client.GetAnalyticsReportSegments(ctx, instanceID, asc.WithAnalyticsReportSegmentsLimit(analyticsMaxLimit))
		}
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if resp.Links.Next == "" {
			break
		}
		next = resp.Links.Next
	}
	return all, nil
}
