package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	reviewscli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/reviews"
)

// ReviewCommand returns the review parent command.
func ReviewCommand() *ffcli.Command {
	return reviewscli.ReviewCommand()
}
