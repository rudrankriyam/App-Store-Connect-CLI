package cmd

import (
	"context"
	"errors"
	"flag"
	"path/filepath"
	"strings"
	"testing"
)

func TestWinBackOffersValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "list missing subscription",
			args:    []string{"win-back-offers", "list"},
			wantErr: "Error: --subscription is required",
		},
		{
			name:    "get missing id",
			args:    []string{"win-back-offers", "get"},
			wantErr: "Error: --id is required",
		},
		{
			name:    "create missing subscription",
			args:    []string{"win-back-offers", "create"},
			wantErr: "Error: --subscription is required",
		},
		{
			name:    "create missing reference-name",
			args:    []string{"win-back-offers", "create", "--subscription", "SUB_ID"},
			wantErr: "Error: --reference-name is required",
		},
		{
			name:    "update missing id",
			args:    []string{"win-back-offers", "update", "--priority", "NORMAL"},
			wantErr: "Error: --id is required",
		},
		{
			name:    "update missing updates",
			args:    []string{"win-back-offers", "update", "--id", "OFFER_ID"},
			wantErr: "Error: at least one update flag is required",
		},
		{
			name:    "delete missing id",
			args:    []string{"win-back-offers", "delete", "--confirm"},
			wantErr: "Error: --id is required",
		},
		{
			name:    "delete missing confirm",
			args:    []string{"win-back-offers", "delete", "--id", "OFFER_ID"},
			wantErr: "Error: --confirm is required",
		},
		{
			name:    "prices missing id",
			args:    []string{"win-back-offers", "prices"},
			wantErr: "Error: --id is required",
		},
		{
			name:    "prices relationships missing id",
			args:    []string{"win-back-offers", "prices-relationships"},
			wantErr: "Error: --id is required",
		},
		{
			name:    "relationships missing subscription",
			args:    []string{"win-back-offers", "relationships"},
			wantErr: "Error: --subscription is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")

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
