package cli

import (
	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"
	"github.com/spf13/cobra"
)

// GetHdacCustomCmd implements custom command especially for Hdac-related contract
func GetHdacCustomCmd(cdc *codec.Codec) *cobra.Command {
	// TODO: Replace as alias of general contract execution
	hdacCustomTxCmd := &cobra.Command{
		Use:                        "hdac",
		Short:                      "Commands for Hdac internal control",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	hdacCustomTxCmd.AddCommand(client.GetCommands(
		// Tx
		GetCmdTransfer(cdc),
		GetCmdBonding(cdc),
		GetCmdUnbonding(cdc),
		GetCmdCreateValidator(cdc),

		// Query
		GetCmdQueryBalance(cdc),
		GetCmdQuery(cdc),
	)...)
	return hdacCustomTxCmd
}
