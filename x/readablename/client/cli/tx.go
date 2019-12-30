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

	"github.com/hdac-io/friday/x/readablename/types"
)

// GetTxCmd handles & routes CLI commands
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	nameserviceTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Readable name service subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nameserviceTxCmd.AddCommand(client.PostCommands(
		GetCmdSetAccount(cdc),
		GetCmdChangeKey(cdc),
	)...)

	return nameserviceTxCmd
}

// GetCmdSetAccount is the CLI command for sending a set account Tx
func GetCmdSetAccount(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "setkey [name] [pubkey]",
		Short: "",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pubkey, err := sdk.GetAccPubKeyBech32(args[1])
			if err != nil {
				return err
			}
			addr := sdk.AccAddress(pubkey.Address())
			straddr := addr.String()
			fmt.Println("Register readable name for ", args[0], " -> ", straddr)

			cliCtx := context.NewCLIContextWithFrom(straddr).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgSetAccount(types.NewName(args[0]), addr, pubkey)
			errValidation := msg.ValidateBasic()
			if errValidation != nil {
				return errValidation
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdChangeKey is the CLI command for changing key
func GetCmdChangeKey(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "changekey [name] [old_public_key] [new_public_key]",
		Short: "",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldpubkey, err := sdk.GetAccPubKeyBech32(args[1])
			if err != nil {
				return err
			}
			oldaddr := sdk.AccAddress(oldpubkey.Address())
			oldstraddr := oldaddr.String()

			newpubkey, err := sdk.GetAccPubKeyBech32(args[2])
			if err != nil {
				return err
			}
			newaddr := sdk.AccAddress(newpubkey.Address())
			newstraddr := newaddr.String()
			fmt.Println("Change key for readable name ", args[0])
			fmt.Println(oldstraddr, " -> ", newstraddr)

			cliCtx := context.NewCLIContextWithFrom(oldstraddr).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgChangeKey(args[0], oldaddr, newaddr, oldpubkey, newpubkey)
			errValidation := msg.ValidateBasic()
			if err != nil {
				return errValidation
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
