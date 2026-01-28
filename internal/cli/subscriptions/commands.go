package subscriptions

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the subscriptions command group.
func Command() *ffcli.Command {
	return SubscriptionsCommand()
}
