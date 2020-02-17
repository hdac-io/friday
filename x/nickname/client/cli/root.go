package cli

import (
	"github.com/spf13/cobra"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"

	"github.com/hdac-io/friday/x/nickname/types"
)

// GetRootCmd handles & routes CLI commands of readable name service
func GetRootCmd(cdc *codec.Codec) *cobra.Command {
	nicknameRootCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Nickname service subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nicknameRootCmd.AddCommand(client.PostCommands(
		// Tx
		GetCmdSetNickname(cdc),
		GetCmdChangeKey(cdc),

		// Query
		GetCmdQueryAddress(cdc),
	)...)

	return nicknameRootCmd
}
