package passtypeids

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the pass-type-ids command group.
func Command() *ffcli.Command {
	return PassTypeIDsCommand()
}
