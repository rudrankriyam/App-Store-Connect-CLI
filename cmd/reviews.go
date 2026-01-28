package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	reviewscli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/reviews"
)

// ReviewsCommand returns the reviews command group.
func ReviewsCommand() *ffcli.Command {
	return reviewscli.ReviewsCommand()
}
