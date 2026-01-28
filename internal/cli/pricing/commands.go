package pricing

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the pricing command group.
func Command() *ffcli.Command {
	return PricingCommand()
}
