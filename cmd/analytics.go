package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	analyticscli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/analytics"
)

// AnalyticsCommand returns the analytics command group.
func AnalyticsCommand() *ffcli.Command {
	return analyticscli.AnalyticsCommand()
}
