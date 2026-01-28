package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	iapcli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/iap"
)

// IAPCommand returns the in-app purchases command group.
func IAPCommand() *ffcli.Command {
	return iapcli.IAPCommand()
}
