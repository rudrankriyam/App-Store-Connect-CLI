package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	subscriptionscli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/subscriptions"
)

// SubscriptionsCommand returns the subscriptions command group.
func SubscriptionsCommand() *ffcli.Command {
	return subscriptionscli.SubscriptionsCommand()
}
