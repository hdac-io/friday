package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/auth/client/utils"

	"github.com/hdac-io/friday/x/nickname/types"
)

// GetTxCmd handles & routes CLI commands
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	nameserviceTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Nickname name service subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nameserviceTxCmd.AddCommand(client.PostCommands(
		GetCmdSetNickname(cdc),
		GetCmdChangeKey(cdc),
	)...)

	return nameserviceTxCmd
}

// GetCmdSetNickname is the CLI command to register nickname from address
func GetCmdSetNickname(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <nickname> --from <from>",
		Short: fmt.Sprintf("Set nickname by address (%sxxxxxx...)", sdk.Bech32MainPrefix),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			addr := cliCtx.GetFromAddress()

			fmt.Println("Register readable name for ", args[0], " -> ", addr.String())

			msg := types.NewMsgSetNickname(types.NewName(args[0]), addr)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")

	return cmd
}

// GetCmdChangeKey is the CLI command for changing key
func GetCmdChangeKey(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-to <nickname> <new_address> --from <old_from>",
		Short: "Change public key mapping of given nickname to address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			oldaddr := cliCtx.GetFromAddress()
			newaddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeKey(args[0], oldaddr, newaddr)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")

	return cmd
}
