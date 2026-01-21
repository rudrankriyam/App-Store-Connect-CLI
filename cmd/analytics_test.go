package cmd

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestAnalyticsSalesValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing vendor",
			args:    []string{"analytics", "sales", "--type", "SALES", "--subtype", "SUMMARY", "--frequency", "DAILY", "--date", "2024-01-20"},
			wantErr: "--vendor is required",
		},
		{
			name:    "missing type",
			args:    []string{"analytics", "sales", "--vendor", "12345678", "--subtype", "SUMMARY", "--frequency", "DAILY", "--date", "2024-01-20"},
			wantErr: "--type is required",
		},
		{
			name:    "missing subtype",
			args:    []string{"analytics", "sales", "--vendor", "12345678", "--type", "SALES", "--frequency", "DAILY", "--date", "2024-01-20"},
			wantErr: "--subtype is required",
		},
		{
			name:    "missing frequency",
			args:    []string{"analytics", "sales", "--vendor", "12345678", "--type", "SALES", "--subtype", "SUMMARY", "--date", "2024-01-20"},
			wantErr: "--frequency is required",
		},
		{
			name:    "missing date",
			args:    []string{"analytics", "sales", "--vendor", "12345678", "--type", "SALES", "--subtype", "SUMMARY", "--frequency", "DAILY"},
			wantErr: "--date is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}

func TestAnalyticsRequestValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing app",
			args:    []string{"analytics", "request", "--access-type", "ONGOING"},
			wantErr: "--app is required",
		},
		{
			name:    "missing access type",
			args:    []string{"analytics", "request", "--app", "APP_ID"},
			wantErr: "--access-type is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}

func TestAnalyticsRequestsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"analytics", "requests"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		err := root.Run(context.Background())
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected ErrHelp, got %v", err)
		}
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "--app is required") {
		t.Fatalf("expected missing app error, got %q", stderr)
	}
}

func TestAnalyticsGetValidationErrors(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"analytics", "get"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		err := root.Run(context.Background())
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected ErrHelp, got %v", err)
		}
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "--request-id is required") {
		t.Fatalf("expected missing request-id error, got %q", stderr)
	}
}

func TestAnalyticsDownloadValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing request id",
			args:    []string{"analytics", "download"},
			wantErr: "--request-id is required",
		},
		{
			name:    "missing instance id",
			args:    []string{"analytics", "download", "--request-id", "11111111-1111-1111-1111-111111111111"},
			wantErr: "--instance-id is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}
