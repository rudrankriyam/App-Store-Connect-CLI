package cmd

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestReviewSubmissionsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "review-submissions list missing app",
			args:    []string{"review-submissions", "list"},
			wantErr: "--app is required",
		},
		{
			name:    "review-submissions get missing id",
			args:    []string{"review-submissions", "get"},
			wantErr: "--id is required",
		},
		{
			name:    "review-submissions create missing app",
			args:    []string{"review-submissions", "create"},
			wantErr: "--app is required",
		},
		{
			name:    "review-submissions submit missing id",
			args:    []string{"review-submissions", "submit", "--confirm"},
			wantErr: "--id is required",
		},
		{
			name:    "review-submissions submit missing confirm",
			args:    []string{"review-submissions", "submit", "--id", "SUBMISSION_123"},
			wantErr: "--confirm is required to submit",
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

func TestReviewItemsValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "review-items list missing submission",
			args:    []string{"review-items", "list"},
			wantErr: "--submission is required",
		},
		{
			name:    "review-items add missing submission",
			args:    []string{"review-items", "add", "--item-type", "appStoreVersions", "--item-id", "VERSION_ID"},
			wantErr: "--submission is required",
		},
		{
			name:    "review-items add missing item-type",
			args:    []string{"review-items", "add", "--submission", "SUBMISSION_ID", "--item-id", "VERSION_ID"},
			wantErr: "--item-type is required",
		},
		{
			name:    "review-items add missing item-id",
			args:    []string{"review-items", "add", "--submission", "SUBMISSION_ID", "--item-type", "appStoreVersions"},
			wantErr: "--item-id is required",
		},
		{
			name:    "review-items remove missing id",
			args:    []string{"review-items", "remove", "--confirm"},
			wantErr: "--id is required",
		},
		{
			name:    "review-items remove missing confirm",
			args:    []string{"review-items", "remove", "--id", "ITEM_ID"},
			wantErr: "--confirm is required to remove",
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

func TestReviewItemsInvalidItemType(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	if err := root.Parse([]string{"review-items", "add", "--submission", "SUBMISSION_ID", "--item-type", "nope", "--item-id", "ITEM_ID"}); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	err := root.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--item-type must be one of") {
		t.Fatalf("expected error mentioning item-type, got %v", err)
	}
}

func TestReviewCommandSubmissionsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "review submissions-list missing app",
			args:    []string{"review", "submissions-list"},
			wantErr: "--app is required",
		},
		{
			name:    "review submissions-get missing id",
			args:    []string{"review", "submissions-get"},
			wantErr: "--id is required",
		},
		{
			name:    "review submissions-create missing app",
			args:    []string{"review", "submissions-create"},
			wantErr: "--app is required",
		},
		{
			name:    "review submissions-submit missing id",
			args:    []string{"review", "submissions-submit", "--confirm"},
			wantErr: "--id is required",
		},
		{
			name:    "review submissions-submit missing confirm",
			args:    []string{"review", "submissions-submit", "--id", "SUBMISSION_123"},
			wantErr: "--confirm is required to submit",
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

func TestReviewCommandItemsValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "review items-list missing submission",
			args:    []string{"review", "items-list"},
			wantErr: "--submission is required",
		},
		{
			name:    "review items-add missing submission",
			args:    []string{"review", "items-add", "--item-type", "appStoreVersions", "--item-id", "VERSION_ID"},
			wantErr: "--submission is required",
		},
		{
			name:    "review items-add missing item-type",
			args:    []string{"review", "items-add", "--submission", "SUBMISSION_ID", "--item-id", "VERSION_ID"},
			wantErr: "--item-type is required",
		},
		{
			name:    "review items-add missing item-id",
			args:    []string{"review", "items-add", "--submission", "SUBMISSION_ID", "--item-type", "appStoreVersions"},
			wantErr: "--item-id is required",
		},
		{
			name:    "review items-remove missing id",
			args:    []string{"review", "items-remove", "--confirm"},
			wantErr: "--id is required",
		},
		{
			name:    "review items-remove missing confirm",
			args:    []string{"review", "items-remove", "--id", "ITEM_ID"},
			wantErr: "--confirm is required to remove",
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

func TestReviewCommandItemsInvalidItemType(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	if err := root.Parse([]string{"review", "items-add", "--submission", "SUBMISSION_ID", "--item-type", "nope", "--item-id", "ITEM_ID"}); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	err := root.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--item-type must be one of") {
		t.Fatalf("expected error mentioning item-type, got %v", err)
	}
}
