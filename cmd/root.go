package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// VersionCommand returns a version subcommand
func VersionCommand(version string) *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "asc version",
		ShortHelp:  "Print version information and exit.",
		UsageFunc:  DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			fmt.Println(version)
			return nil
		},
	}
}

// RootCommand returns the root command
func RootCommand(version string) *ffcli.Command {
	root := &ffcli.Command{
		Name:       "asc",
		ShortUsage: "asc <subcommand> [flags]",
		ShortHelp:  "A fast, AI-agent friendly CLI for App Store Connect.",
		LongHelp:   "ASC is a lightweight CLI for App Store Connect. Built for developers and AI agents.",
		FlagSet:    flag.NewFlagSet("asc", flag.ExitOnError),
		UsageFunc:  DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AuthCommand(),
			FeedbackCommand(),
			CrashesCommand(),
			ReviewsCommand(),
			AnalyticsCommand(),
			FinanceCommand(),
			AppsCommand(),
			TestFlightCommand(),
			BuildsCommand(),
			VersionsCommand(),
			PreReleaseVersionsCommand(),
			LocalizationsCommand(),
			AssetsCommand(),
			BuildLocalizationsCommand(),
			BetaGroupsCommand(),
			BetaTestersCommand(),
			SandboxCommand(),
			SubmitCommand(),
			XcodeCloudCommand(),
			VersionCommand(version),
		},
	}

	versionFlag := root.FlagSet.Bool("version", false, "Print version and exit")

	root.Exec = func(ctx context.Context, args []string) error {
		if *versionFlag {
			fmt.Fprintln(os.Stdout, version)
			return nil
		}
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		}
		return flag.ErrHelp
	}

	return root
}
