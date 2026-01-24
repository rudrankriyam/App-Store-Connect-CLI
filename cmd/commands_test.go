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

func TestBuildsGroupValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "builds add-groups missing build",
			args:    []string{"builds", "add-groups"},
			wantErr: "Error: --build is required",
		},
		{
			name:    "builds add-groups missing group",
			args:    []string{"builds", "add-groups", "--build", "BUILD_123"},
			wantErr: "Error: --group is required",
		},
		{
			name:    "builds remove-groups missing build",
			args:    []string{"builds", "remove-groups"},
			wantErr: "Error: --build is required",
		},
		{
			name:    "builds remove-groups missing group",
			args:    []string{"builds", "remove-groups", "--build", "BUILD_123"},
			wantErr: "Error: --group is required",
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
			name:    "beta-testers add-groups missing id",
			args:    []string{"beta-testers", "add-groups", "--group", "GROUP_ID"},
			wantErr: "--id is required",
		},
		{
			name:    "beta-testers add-groups missing group",
			args:    []string{"beta-testers", "add-groups", "--id", "TESTER_ID"},
			wantErr: "--group is required",
		},
		{
			name:    "beta-testers remove-groups missing id",
			args:    []string{"beta-testers", "remove-groups", "--group", "GROUP_ID"},
			wantErr: "--id is required",
		},
		{
			name:    "beta-testers remove-groups missing group",
			args:    []string{"beta-testers", "remove-groups", "--id", "TESTER_ID"},
			wantErr: "--group is required",
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
		{
			name:    "beta-testers get missing id",
			args:    []string{"beta-testers", "get"},
			wantErr: "--id is required",
		},
		{
			name:    "beta-groups get missing id",
			args:    []string{"beta-groups", "get"},
			wantErr: "--id is required",
		},
		{
			name:    "beta-groups update missing id",
			args:    []string{"beta-groups", "update"},
			wantErr: "--id is required",
		},
		{
			name:    "beta-groups update missing update flags",
			args:    []string{"beta-groups", "update", "--id", "GROUP_ID"},
			wantErr: "at least one update flag is required",
		},
		{
			name:    "beta-groups update public-link-limit out of range",
			args:    []string{"beta-groups", "update", "--id", "GROUP_ID", "--public-link-limit", "50000"},
			wantErr: "--public-link-limit must be between 1 and 10000",
		},
		{
			name:    "beta-groups add-testers missing group",
			args:    []string{"beta-groups", "add-testers"},
			wantErr: "--group is required",
		},
		{
			name:    "beta-groups add-testers missing tester",
			args:    []string{"beta-groups", "add-testers", "--group", "GROUP_ID"},
			wantErr: "--tester is required",
		},
		{
			name:    "beta-groups remove-testers missing group",
			args:    []string{"beta-groups", "remove-testers"},
			wantErr: "--group is required",
		},
		{
			name:    "beta-groups remove-testers missing tester",
			args:    []string{"beta-groups", "remove-testers", "--group", "GROUP_ID"},
			wantErr: "--tester is required",
		},
		{
			name:    "beta-groups delete missing id",
			args:    []string{"beta-groups", "delete"},
			wantErr: "--id is required",
		},
		{
			name:    "beta-groups delete missing confirm",
			args:    []string{"beta-groups", "delete", "--id", "GROUP_ID"},
			wantErr: "--confirm is required to delete",
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

func TestTestFlightAppsValidationErrors(t *testing.T) {
	t.Setenv("ASC_KEY_ID", "")
	t.Setenv("ASC_ISSUER_ID", "")
	t.Setenv("ASC_PRIVATE_KEY_PATH", "")

	tests := []struct {
		name     string
		args     []string
		wantErr  string
		wantHelp bool
	}{
		{
			name:     "testflight apps list missing auth",
			args:     []string{"testflight", "apps", "list"},
			wantErr:  "missing authentication",
			wantHelp: false,
		},
		{
			name:     "testflight apps get missing id",
			args:     []string{"testflight", "apps", "get"},
			wantErr:  "--app is required",
			wantHelp: true,
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
				if test.wantHelp {
					if !errors.Is(err, flag.ErrHelp) {
						t.Fatalf("expected ErrHelp, got %v", err)
					}
				} else {
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if !strings.Contains(err.Error(), test.wantErr) {
						t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
					}
				}
			})

			if test.wantHelp {
				if stdout != "" {
					t.Fatalf("expected empty stdout, got %q", stdout)
				}
				if !strings.Contains(stderr, test.wantErr) {
					t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
				}
			}
		})
	}
}

func TestTestFlightSyncValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "testflight sync pull missing app",
			args:    []string{"testflight", "sync", "pull", "--output", "./testflight.yaml"},
			wantErr: "--app is required",
		},
		{
			name:    "testflight sync pull missing output",
			args:    []string{"testflight", "sync", "pull", "--app", "APP_ID"},
			wantErr: "--output is required",
		},
		{
			name:    "testflight sync pull build filter without include",
			args:    []string{"testflight", "sync", "pull", "--app", "APP_ID", "--output", "./testflight.yaml", "--build", "BUILD_ID"},
			wantErr: "--build requires --include-builds",
		},
		{
			name:    "testflight sync pull tester filter without include",
			args:    []string{"testflight", "sync", "pull", "--app", "APP_ID", "--output", "./testflight.yaml", "--tester", "tester@example.com"},
			wantErr: "--tester requires --include-testers",
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

func TestParseCommaSeparatedIDs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "single id",
			input: "tester-1",
			want:  []string{"tester-1"},
		},
		{
			name:  "comma separated",
			input: "tester-1, tester-2, tester-3",
			want:  []string{"tester-1", "tester-2", "tester-3"},
		},
		{
			name:  "drops empty entries",
			input: "tester-1,, ,tester-2",
			want:  []string{"tester-1", "tester-2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseCommaSeparatedIDs(test.input)
			if len(got) != len(test.want) {
				t.Fatalf("expected %d ids, got %d", len(test.want), len(got))
			}
			for i, want := range test.want {
				if got[i] != want {
					t.Fatalf("expected %q at index %d, got %q", want, i, got[i])
				}
			}
		})
	}
}

func TestBetaTestersListAcceptsBuildFilter(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	if err := root.Parse([]string{"beta-testers", "list", "--app", "X", "--build", "Y"}); err != nil {
		t.Fatalf("parse error: %v", err)
	}
}

func TestLocalizationsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "localizations list missing version",
			args:    []string{"localizations", "list"},
			wantErr: "--version is required",
		},
		{
			name:    "localizations list missing app for app-info",
			args:    []string{"localizations", "list", "--type", "app-info"},
			wantErr: "--app is required",
		},
		{
			name:    "localizations download missing version",
			args:    []string{"localizations", "download"},
			wantErr: "--version is required",
		},
		{
			name:    "localizations download missing app for app-info",
			args:    []string{"localizations", "download", "--type", "app-info"},
			wantErr: "--app is required",
		},
		{
			name:    "localizations upload missing path",
			args:    []string{"localizations", "upload", "--version", "VERSION_ID"},
			wantErr: "--path is required",
		},
		{
			name:    "localizations upload missing version",
			args:    []string{"localizations", "upload", "--path", "localizations"},
			wantErr: "--version is required",
		},
		{
			name:    "localizations upload missing app for app-info",
			args:    []string{"localizations", "upload", "--type", "app-info", "--path", "localizations"},
			wantErr: "--app is required",
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

func TestAssetsValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "assets screenshots list missing localization",
			args:    []string{"assets", "screenshots", "list"},
			wantErr: "--version-localization is required",
		},
		{
			name:    "assets screenshots upload missing localization",
			args:    []string{"assets", "screenshots", "upload", "--path", "./screenshots", "--device-type", "IPHONE_65"},
			wantErr: "--version-localization is required",
		},
		{
			name:    "assets screenshots upload missing path",
			args:    []string{"assets", "screenshots", "upload", "--version-localization", "LOC_ID", "--device-type", "IPHONE_65"},
			wantErr: "--path is required",
		},
		{
			name:    "assets screenshots upload missing device type",
			args:    []string{"assets", "screenshots", "upload", "--version-localization", "LOC_ID", "--path", "./screenshots"},
			wantErr: "--device-type is required",
		},
		{
			name:    "assets screenshots delete missing id",
			args:    []string{"assets", "screenshots", "delete"},
			wantErr: "--id is required",
		},
		{
			name:    "assets screenshots delete missing confirm",
			args:    []string{"assets", "screenshots", "delete", "--id", "SCREENSHOT_ID"},
			wantErr: "--confirm is required to delete",
		},
		{
			name:    "assets previews list missing localization",
			args:    []string{"assets", "previews", "list"},
			wantErr: "--version-localization is required",
		},
		{
			name:    "assets previews upload missing localization",
			args:    []string{"assets", "previews", "upload", "--path", "./previews", "--device-type", "IPHONE_65"},
			wantErr: "--version-localization is required",
		},
		{
			name:    "assets previews upload missing path",
			args:    []string{"assets", "previews", "upload", "--version-localization", "LOC_ID", "--device-type", "IPHONE_65"},
			wantErr: "--path is required",
		},
		{
			name:    "assets previews upload missing device type",
			args:    []string{"assets", "previews", "upload", "--version-localization", "LOC_ID", "--path", "./previews"},
			wantErr: "--device-type is required",
		},
		{
			name:    "assets previews delete missing id",
			args:    []string{"assets", "previews", "delete"},
			wantErr: "--id is required",
		},
		{
			name:    "assets previews delete missing confirm",
			args:    []string{"assets", "previews", "delete", "--id", "PREVIEW_ID"},
			wantErr: "--confirm is required to delete",
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

func TestBuildLocalizationsValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "build-localizations list missing build",
			args:    []string{"build-localizations", "list"},
			wantErr: "--build is required",
		},
		{
			name:    "build-localizations create missing locale",
			args:    []string{"build-localizations", "create", "--build", "BUILD_ID"},
			wantErr: "--locale is required",
		},
		{
			name:    "build-localizations delete missing confirm",
			args:    []string{"build-localizations", "delete", "--id", "LOC_ID"},
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
			name:    "create missing confirm",
			args:    []string{"submit", "create", "--app", "APP_123", "--version", "1.0.0", "--build", "BUILD_123"},
			wantErr: "Error: --confirm is required",
		},
		{
			name:    "create missing build",
			args:    []string{"submit", "create", "--app", "APP_123", "--version", "1.0.0", "--confirm"},
			wantErr: "Error: --build is required",
		},
		{
			name:    "create missing version",
			args:    []string{"submit", "create", "--app", "APP_123", "--build", "BUILD_123", "--confirm"},
			wantErr: "Error: --version or --version-id is required",
		},
		{
			name:    "status missing id",
			args:    []string{"submit", "status"},
			wantErr: "Error: --id or --version-id is required",
		},
		{
			name:    "cancel missing confirm",
			args:    []string{"submit", "cancel", "--id", "SUBMIT_123"},
			wantErr: "Error: --confirm is required",
		},
		{
			name:    "cancel missing id",
			args:    []string{"submit", "cancel", "--confirm"},
			wantErr: "Error: --id or --version-id is required",
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

func TestSubmitValidationConflicts(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	err := root.Parse([]string{"submit", "create", "--app", "APP_123", "--version", "1.0.0", "--version-id", "VERSION_123", "--build", "BUILD_123", "--confirm"})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if err := root.Run(context.Background()); err == nil {
		t.Fatalf("expected error for conflicting flags")
	}
}

func TestVersionsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "list missing app",
			args:    []string{"versions", "list"},
			wantErr: "Error: --app is required",
		},
		{
			name:    "get missing version id",
			args:    []string{"versions", "get"},
			wantErr: "Error: --version-id is required",
		},
		{
			name:    "attach missing version id",
			args:    []string{"versions", "attach-build", "--build", "BUILD_123"},
			wantErr: "Error: --version-id is required",
		},
		{
			name:    "attach missing build",
			args:    []string{"versions", "attach-build", "--version-id", "VERSION_123"},
			wantErr: "Error: --build is required",
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

func TestPreReleaseVersionsValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "pre-release-versions list missing app",
			args:    []string{"pre-release-versions", "list"},
			wantErr: "Error: --app is required",
		},
		{
			name:    "pre-release-versions get missing id",
			args:    []string{"pre-release-versions", "get"},
			wantErr: "Error: --id is required",
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

func TestXcodeCloudValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "xcode-cloud run missing workflow",
			args:    []string{"xcode-cloud", "run", "--branch", "main"},
			wantErr: "--workflow or --workflow-id is required",
		},
		{
			name:    "xcode-cloud run missing branch",
			args:    []string{"xcode-cloud", "run", "--workflow-id", "WF_ID"},
			wantErr: "--branch or --git-reference-id is required",
		},
		{
			name:    "xcode-cloud run workflow by name without app",
			args:    []string{"xcode-cloud", "run", "--workflow", "CI", "--branch", "main"},
			wantErr: "--app is required when using --workflow",
		},
		{
			name:    "xcode-cloud status missing run-id",
			args:    []string{"xcode-cloud", "status"},
			wantErr: "--run-id is required",
		},
		{
			name:    "xcode-cloud workflows missing app",
			args:    []string{"xcode-cloud", "workflows"},
			wantErr: "--app is required",
		},
		{
			name:    "xcode-cloud build-runs missing workflow-id",
			args:    []string{"xcode-cloud", "build-runs"},
			wantErr: "--workflow-id is required",
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

func TestXcodeCloudMutualExclusiveFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "xcode-cloud run workflow and workflow-id are mutually exclusive",
			args:    []string{"xcode-cloud", "run", "--workflow", "CI", "--workflow-id", "WF_ID", "--branch", "main"},
			wantErr: "--workflow and --workflow-id are mutually exclusive",
		},
		{
			name:    "xcode-cloud run branch and git-reference-id are mutually exclusive",
			args:    []string{"xcode-cloud", "run", "--workflow-id", "WF_ID", "--branch", "main", "--git-reference-id", "REF_ID"},
			wantErr: "--branch and --git-reference-id are mutually exclusive",
		},
		{
			name:    "xcode-cloud run invalid poll-interval",
			args:    []string{"xcode-cloud", "run", "--workflow-id", "WF_ID", "--branch", "main", "--wait", "--poll-interval", "0s"},
			wantErr: "--poll-interval must be greater than 0",
		},
		{
			name:    "xcode-cloud status invalid timeout",
			args:    []string{"xcode-cloud", "status", "--run-id", "RUN_ID", "--timeout", "-1s"},
			wantErr: "--timeout must be greater than or equal to 0",
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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
				}
			})

			// These errors come from Exec, not from validation that returns ErrHelp
			_ = stdout
			_ = stderr
		})
	}
}
