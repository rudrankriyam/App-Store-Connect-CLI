package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	pricingcli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/pricing"
)

// PricingCommand returns the pricing command group.
func PricingCommand() *ffcli.Command {
	return pricingcli.PricingCommand()
}
