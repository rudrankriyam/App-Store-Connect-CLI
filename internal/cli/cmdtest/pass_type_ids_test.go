package cmdtest

import (
	"context"
	"errors"
	"flag"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestPassTypeIDsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "pass-type-ids list missing pass type id for certificates list",
			args:    []string{"pass-type-ids", "certificates", "list"},
			wantErr: "--pass-type-id is required",
		},
		{
			name:    "pass-type-ids certificates get missing pass type id",
			args:    []string{"pass-type-ids", "certificates", "get"},
			wantErr: "--pass-type-id is required",
		},
		{
			name:    "pass-type-ids get missing pass type id",
			args:    []string{"pass-type-ids", "get"},
			wantErr: "--pass-type-id is required",
		},
		{
			name:    "pass-type-ids create missing identifier",
			args:    []string{"pass-type-ids", "create", "--name", "Example"},
			wantErr: "--identifier is required",
		},
		{
			name:    "pass-type-ids create missing name",
			args:    []string{"pass-type-ids", "create", "--identifier", "pass.com.example"},
			wantErr: "--name is required",
		},
		{
			name:    "pass-type-ids update missing pass type id",
			args:    []string{"pass-type-ids", "update", "--name", "Updated"},
			wantErr: "--pass-type-id is required",
		},
		{
			name:    "pass-type-ids update missing name",
			args:    []string{"pass-type-ids", "update", "--pass-type-id", "PASS_ID"},
			wantErr: "--name is required",
		},
		{
			name:    "pass-type-ids delete missing pass type id",
			args:    []string{"pass-type-ids", "delete", "--confirm"},
			wantErr: "--pass-type-id is required",
		},
		{
			name:    "pass-type-ids delete missing confirm",
			args:    []string{"pass-type-ids", "delete", "--pass-type-id", "PASS_ID"},
			wantErr: "--confirm is required",
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
