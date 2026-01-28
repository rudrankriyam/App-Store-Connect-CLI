package iap

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the iap command group.
func Command() *ffcli.Command {
	return IAPCommand()
}
