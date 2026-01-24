package cmd

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// SandboxCreateCommand returns the sandbox create subcommand.
func SandboxCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ExitOnError)

	email := fs.String("email", "", "Tester email address")
	firstName := fs.String("first-name", "", "Tester first name")
	lastName := fs.String("last-name", "", "Tester last name")
	password := fs.String("password", "", "Tester password (8+ chars, uppercase, lowercase, number)")
	passwordStdin := fs.Bool("password-stdin", false, "Read tester password from stdin")
	confirmPassword := fs.String("confirm-password", "", "Confirm password (must match --password)")
	secretQuestion := fs.String("secret-question", "", "Secret question (6+ chars)")
	secretAnswer := fs.String("secret-answer", "", "Secret answer (6+ chars)")
	secretAnswerStdin := fs.Bool("secret-answer-stdin", false, "Read secret answer from stdin")
	birthDate := fs.String("birth-date", "", "Birth date (YYYY-MM-DD)")
	territory := fs.String("territory", "", "App Store territory code (e.g., USA, JPN)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc sandbox create [flags]",
		ShortHelp:  "Create a sandbox tester.",
		LongHelp: `Create a new sandbox tester account for in-app purchase testing.

Examples:
  asc sandbox create --email "tester@example.com" --first-name "Test" --last-name "User" --password "Passwordtest1" --confirm-password "Passwordtest1" --secret-question "Question" --secret-answer "Answer" --birth-date "1980-03-01" --territory "USA"
  echo "Passwordtest1" | asc sandbox create --email "tester@example.com" --first-name "Test" --last-name "User" --password-stdin --secret-question "Question" --secret-answer "Answer" --birth-date "1980-03-01" --territory "USA"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --email is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*firstName) == "" {
				fmt.Fprintln(os.Stderr, "Error: --first-name is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*lastName) == "" {
				fmt.Fprintln(os.Stderr, "Error: --last-name is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*secretQuestion) == "" {
				fmt.Fprintln(os.Stderr, "Error: --secret-question is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*birthDate) == "" {
				fmt.Fprintln(os.Stderr, "Error: --birth-date is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*territory) == "" {
				fmt.Fprintln(os.Stderr, "Error: --territory is required")
				return flag.ErrHelp
			}

			if *passwordStdin && *secretAnswerStdin {
				return fmt.Errorf("sandbox create: --password-stdin and --secret-answer-stdin cannot both be set")
			}

			readStdinSecret := func(flagName string) (string, error) {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return "", fmt.Errorf("%s: failed to read stdin: %w", flagName, err)
				}
				value := strings.TrimSpace(string(data))
				if value == "" {
					return "", fmt.Errorf("%s requires a non-empty value from stdin", flagName)
				}
				return value, nil
			}

			passwordValue := strings.TrimSpace(*password)
			confirmValue := strings.TrimSpace(*confirmPassword)
			secretAnswerValue := strings.TrimSpace(*secretAnswer)

			if *passwordStdin {
				if passwordValue != "" {
					return fmt.Errorf("sandbox create: --password and --password-stdin are mutually exclusive")
				}
				value, err := readStdinSecret("--password-stdin")
				if err != nil {
					return fmt.Errorf("sandbox create: %w", err)
				}
				passwordValue = value
				if confirmValue == "" {
					confirmValue = passwordValue
				}
			}

			if *secretAnswerStdin {
				if secretAnswerValue != "" {
					return fmt.Errorf("sandbox create: --secret-answer and --secret-answer-stdin are mutually exclusive")
				}
				value, err := readStdinSecret("--secret-answer-stdin")
				if err != nil {
					return fmt.Errorf("sandbox create: %w", err)
				}
				secretAnswerValue = value
			}

			if passwordValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --password is required (or use --password-stdin)")
				return flag.ErrHelp
			}
			if confirmValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --confirm-password is required")
				return flag.ErrHelp
			}
			if secretAnswerValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --secret-answer is required (or use --secret-answer-stdin)")
				return flag.ErrHelp
			}

			if err := validateSandboxEmail(*email); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			if err := validateSandboxPassword(passwordValue); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			if confirmValue != passwordValue {
				return fmt.Errorf("sandbox create: --confirm-password must match --password")
			}
			if err := validateSandboxSecret("--secret-question", *secretQuestion); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			if err := validateSandboxSecret("--secret-answer", secretAnswerValue); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}

			normalizedBirthDate, err := normalizeSandboxBirthDate(*birthDate)
			if err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			normalizedTerritory, err := normalizeSandboxTerritory(*territory)
			if err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return err
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			attrs := asc.SandboxTesterCreateAttributes{
				FirstName:         strings.TrimSpace(*firstName),
				LastName:          strings.TrimSpace(*lastName),
				Email:             strings.TrimSpace(*email),
				Password:          passwordValue,
				ConfirmPassword:   confirmValue,
				SecretQuestion:    strings.TrimSpace(*secretQuestion),
				SecretAnswer:      secretAnswerValue,
				BirthDate:         normalizedBirthDate,
				AppStoreTerritory: normalizedTerritory,
			}

			resp, err := client.CreateSandboxTester(requestCtx, attrs)
			if err != nil {
				if asc.IsNotFound(err) {
					return fmt.Errorf("sandbox create: sandbox tester creation is not available via the App Store Connect API for this account")
				}
				return fmt.Errorf("sandbox create: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}
