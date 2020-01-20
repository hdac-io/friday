package cli

import (
	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/readablename/types"
	"github.com/spf13/cobra"
)

// GetRootCmd handles & routes CLI commands of readable name service
func GetRootCmd(cdc *codec.Codec) *cobra.Command {
	readablenameRootCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Readable name service subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	readablenameRootCmd.AddCommand(client.PostCommands(
		// Tx
		GetCmdSetAccountFromBech32PubKey(cdc),
		GetCmdSetAccountFromSecp256k1PubKey(cdc),
		GetCmdChangeKeyFromBech32PubKey(cdc),
		GetCmdChangeKeyFromSecp256k1PubKey(cdc),

		// Query
		GetCmdQueryUnitAccount(cdc),
	)...)

	return readablenameRootCmd
}
