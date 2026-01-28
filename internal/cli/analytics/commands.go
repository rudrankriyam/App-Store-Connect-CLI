package analytics

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the analytics command group.
func Command() *ffcli.Command {
	return AnalyticsCommand()
}
