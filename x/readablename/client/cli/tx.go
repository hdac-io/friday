package cli

import (
	"github.com/spf13/cobra"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/auth/client/utils"

	"github.com/hdac-io/friday/x/readablename/types"
)

// GetTxCmd handles & routes CLI commands
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	nameserviceTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Nameserver post Tx subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nameserviceTxCmd.AddCommand(client.PostCommands(
		GetCmdSetAccount(cdc),
		GetCmdChangeKey(cdc),
		GetCmdAddrCheck(cdc),
	)...)

	return nameserviceTxCmd
}

// GetCmdSetAccount is the CLI command for sending a set account Tx
func GetCmdSetAccount(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "setaccount [name] [address]",
		Short: "",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			//cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			addr, _ := sdk.AccAddressFromBech32(args[1])
			msg := types.NewMsgSetAccount(args[0], addr)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdChangeKey is the CLI command for changing key
func GetCmdChangeKey(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "changekey [name] [old private key] [new private key]",
		Short: "",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			oldaddr, _ := sdk.AccAddressFromBech32(args[1])
			newaddr, _ := sdk.AccAddressFromBech32(args[2])
			msg := types.NewMsgChangeKey(args[0], oldaddr, newaddr)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdAddrCheck is the CLI command for changing key
func GetCmdAddrCheck(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "addrcheck [name] [address]",
		Short: "",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			addr, _ := sdk.AccAddressFromBech32(args[1])
			msg := types.NewMsgAddrCheck(args[0], addr)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
