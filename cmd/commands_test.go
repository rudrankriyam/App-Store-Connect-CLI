package cmd

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		_ = rOut.Close()
		outC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		_ = rErr.Close()
		errC <- buf.String()
	}()

	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		_ = wOut.Close()
		_ = wErr.Close()
	}()

	fn()

	_ = wOut.Close()
	_ = wErr.Close()

	stdout := <-outC
	stderr := <-errC

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return stdout, stderr
}

func TestVersionSubcommandPrintsVersion(t *testing.T) {
	root := RootCommand("1.2.3")

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"version"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stdout != "1.2.3\n" {
		t.Fatalf("expected stdout to be version, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestVersionFlagPrintsVersion(t *testing.T) {
	root := RootCommand("1.2.3")

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"--version"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stdout != "1.2.3\n" {
		t.Fatalf("expected stdout to be version, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestUnknownCommandPrintsHelpError(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"nope"}); err != nil {
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
	if !strings.Contains(stderr, "Unknown command: nope") {
		t.Fatalf("expected unknown command message, got %q", stderr)
	}
}

func TestBuildsInfoRequiresBuildID(t *testing.T) {
	root := RootCommand("1.2.3")

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"builds", "info"}); err != nil {
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
	if !strings.Contains(stderr, "--build is required") {
		t.Fatalf("expected missing build error, got %q", stderr)
	}
}

func TestBuildsExpireRequiresBuildID(t *testing.T) {
	root := RootCommand("1.2.3")

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"builds", "expire"}); err != nil {
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
	if !strings.Contains(stderr, "--build is required") {
		t.Fatalf("expected missing build error, got %q", stderr)
	}
}

func TestBetaManagementValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "beta-groups list missing app",
			args:    []string{"beta-groups", "list"},
			wantErr: "--app is required",
		},
		{
			name:    "beta-groups create missing app",
			args:    []string{"beta-groups", "create", "--name", "Beta"},
			wantErr: "--app is required",
		},
		{
			name:    "beta-groups create missing name",
			args:    []string{"beta-groups", "create", "--app", "APP_ID"},
			wantErr: "--name is required",
		},
		{
			name:    "beta-testers list missing app",
			args:    []string{"beta-testers", "list"},
			wantErr: "--app is required",
		},
		{
			name:    "beta-testers add missing app",
			args:    []string{"beta-testers", "add", "--email", "tester@example.com", "--group", "Beta"},
			wantErr: "--app is required",
		},
		{
			name:    "beta-testers add missing email",
			args:    []string{"beta-testers", "add", "--app", "APP_ID", "--group", "Beta"},
			wantErr: "--email is required",
		},
		{
			name:    "beta-testers add missing group",
			args:    []string{"beta-testers", "add", "--app", "APP_ID", "--email", "tester@example.com"},
			wantErr: "--group is required",
		},
		{
			name:    "beta-testers remove missing app",
			args:    []string{"beta-testers", "remove", "--email", "tester@example.com"},
			wantErr: "--app is required",
		},
		{
			name:    "beta-testers remove missing email",
			args:    []string{"beta-testers", "remove", "--app", "APP_ID"},
			wantErr: "--email is required",
		},
		{
			name:    "beta-testers invite missing app",
			args:    []string{"beta-testers", "invite", "--email", "tester@example.com"},
			wantErr: "--app is required",
		},
		{
			name:    "beta-testers invite missing email",
			args:    []string{"beta-testers", "invite", "--app", "APP_ID"},
			wantErr: "--email is required",
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

func TestBuildsUploadValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing app",
			args:    []string{"builds", "upload", "--ipa", "app.ipa", "--version", "1.0.0", "--build-number", "123"},
			wantErr: "Error: --app is required",
		},
		{
			name:    "missing ipa",
			args:    []string{"builds", "upload", "--app", "APP_123", "--version", "1.0.0", "--build-number", "123"},
			wantErr: "Error: --ipa is required",
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

func TestSubmitValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing version",
			args:    []string{"submit", "--confirm"},
			wantErr: "Error: --version is required",
		},
		{
			name:    "missing confirm",
			args:    []string{"submit", "--version", "VERSION_123"},
			wantErr: "Error: --confirm is required",
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
