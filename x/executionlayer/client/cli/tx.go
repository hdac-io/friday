package cli

import (
	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/spf13/cobra"
)

// GetExecutionLayerTxCmd controls Tx request of CLI interface
func GetExecutionLayerTxCmd(cdc *codec.Codec) *cobra.Command {
	executionlayerTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Tx commands for execution layer",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	executionlayerTxCmd.AddCommand(client.GetCommands(
	//GetCmdQueryBalance(cdc),
	)...)
	return executionlayerTxCmd
}

// The code below is a pattern of sending Tx
// You may start from the example first

// // GetCmdChangeKey is the CLI command for changing key
// func GetCmdChangeKey(cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "changekey [name] [old private key] [new private key]",
// 		Short: "",
// 		Args:  cobra.ExactArgs(3), // # of arguments
// 		RunE: func(cmd *cobra.Command, args []string) error {
//			// Make context and tx builder
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)
//			// TxBuilder is defined in auth package so you may use it.
// 			// Don't have to find NewTxBuilderFromCLI in your module
// 			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

//			// Parse arguments
// 			oldaddr, _ := sdk.AccAddressFromBech32(args[1])
// 			newaddr, _ := sdk.AccAddressFromBech32(args[2])
//			// Build messages
// 			msg := types.NewMsgChangeKey(args[0], oldaddr, newaddr)
// 			err := msg.ValidateBasic()
// 			if err != nil {
// 				return err
// 			}

//			// Broadcast message in Tx
// 			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
// 		},
// 	}
// }
