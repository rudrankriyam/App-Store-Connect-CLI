package merchantids

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the merchant-ids command group.
func Command() *ffcli.Command {
	return MerchantIDsCommand()
}
