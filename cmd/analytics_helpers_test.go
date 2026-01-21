package cmd

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

func TestResolveReportOutputPaths_Decompress(t *testing.T) {
	compressed, decompressed := resolveReportOutputPaths("report.tsv.gz", "default.tsv.gz", ".tsv", true)
	if compressed != "report.tsv.gz" {
		t.Fatalf("expected compressed path report.tsv.gz, got %q", compressed)
	}
	if decompressed != "report.tsv" {
		t.Fatalf("expected decompressed path report.tsv, got %q", decompressed)
	}

	compressed, decompressed = resolveReportOutputPaths("report.tsv", "default.tsv.gz", ".tsv", true)
	if compressed != "report.tsv.gz" {
		t.Fatalf("expected compressed path report.tsv.gz, got %q", compressed)
	}
	if decompressed != "report.tsv" {
		t.Fatalf("expected decompressed path report.tsv, got %q", decompressed)
	}

	compressed, decompressed = resolveReportOutputPaths("report", "default.tsv.gz", ".tsv", true)
	if compressed != "report" {
		t.Fatalf("expected compressed path report, got %q", compressed)
	}
	if decompressed != "report.tsv" {
		t.Fatalf("expected decompressed path report.tsv, got %q", decompressed)
	}
}

func TestNormalizeReportDate_MonthlyValidation(t *testing.T) {
	_, err := normalizeReportDate("2024-01-02", asc.SalesReportFrequencyMonthly)
	if err == nil {
		t.Fatal("expected error for non-first day monthly date")
	}
}

func TestNormalizeReportDate_MonthlyFormat(t *testing.T) {
	date, err := normalizeReportDate("2024-01", asc.SalesReportFrequencyMonthly)
	if err != nil {
		t.Fatalf("expected monthly date to parse, got %v", err)
	}
	if date != "2024-01" {
		t.Fatalf("expected date to be 2024-01, got %q", date)
	}
}

func TestNormalizeReportDate_YearlyFormat(t *testing.T) {
	date, err := normalizeReportDate("2024", asc.SalesReportFrequencyYearly)
	if err != nil {
		t.Fatalf("expected yearly date to parse, got %v", err)
	}
	if date != "2024" {
		t.Fatalf("expected date to be 2024, got %q", date)
	}
}

func TestDecompressGzipFile(t *testing.T) {
	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source.tsv.gz")
	dest := filepath.Join(tempDir, "dest.tsv")

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte("hello")); err != nil {
		t.Fatalf("failed to write gzip: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("failed to close gzip: %v", err)
	}
	if err := os.WriteFile(source, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to write source gzip: %v", err)
	}

	size, err := decompressGzipFile(source, dest)
	if err != nil {
		t.Fatalf("decompressGzipFile() error: %v", err)
	}
	if size == 0 {
		t.Fatalf("expected non-zero decompressed size")
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected decompressed content to be hello, got %q", string(data))
	}
}
