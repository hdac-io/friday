package cli

import (
	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/spf13/cobra"
)

// GetExecutionLayerCmd controls Tx request of CLI interface
func GetExecutionLayerCmd(cdc *codec.Codec) *cobra.Command {
	executionlayerTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Tx commands for execution layer",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	executionlayerTxCmd.AddCommand(client.GetCommands(
		// Tx
		GetCmdTransfer(cdc),
		GetCmdBonding(cdc),
		GetCmdUnbonding(cdc),

		// Query
		GetCmdQueryBalance(cdc),
		GetCmdQueryBalanceWithBlockHash(cdc),
		GetCmdQuery(cdc),
		GetCmdQueryWithHash(cdc),
	)...)
	return executionlayerTxCmd
}
