package reviews

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the reviews command group.
func Command() *ffcli.Command {
	return ReviewsCommand()
}
