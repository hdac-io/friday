package cli

import (
	"fmt"
	"strconv"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"
	cliutil "github.com/hdac-io/friday/x/executionlayer/client/util"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetCmdContractRun(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <type> <wasm-path>|<uref>|<name>|<hash> <argument> <fee> <gas_price> --from <from>",
		Short: "Run contract",
		Long: "Run contract\n" +
			"There are 4 types of contract run. ('wasm', 'uref', 'name', 'hash)",
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			keyInfo, err := cliutil.GetLocalWalletInfo(valueFromFromFlag, kb, cdc, cliCtx)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			fromAddr := keyInfo.GetAddress()

			sessionType := cliutil.GetContractType(args[0])
			var sessionCode []byte
			switch sessionType {
			case util.WASM:
				sessionCode = util.LoadWasmFile(args[1])
			case util.HASH:
			case util.UREF:
				sessionCode = util.DecodeHexString(args[1])
			case util.NAME:
				sessionCode = []byte(args[1])
			default:
				return fmt.Errorf("type must be one of wasm, name, uref, or hash")
			}

			gasPrice, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgExecute(
				"dummyAddress",
				fromAddr,
				sessionType,
				sessionCode,
				args[2],
				args[3],
				gasPrice,
			)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdTransfer is the CLI command for transfer
func GetCmdTransfer(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-to <recipient_nickname>|<address> <amount> <fee> <gas_price> --from <from>",
		Short: "Transfer Hdac token",
		Long:  "Transfer Hdac token",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Parse nickname of address
			var recipentAddr sdk.AccAddress
			recipentAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				recipentAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, args[0])
				if err != nil {
					return fmt.Errorf("no nickname mapping of %s", args[0])
				}
			}

			// Numbers parsing
			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			keyInfo, err := cliutil.GetLocalWalletInfo(valueFromFromFlag, kb, cdc, cliCtx)
			if err != nil {
				return err
			}
			fromAddr := keyInfo.GetAddress()

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgTransfer("dummyAddress", fromAddr, recipentAddr, args[1], args[2], gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdBonding is the CLI command for bonding
func GetCmdBonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond --from <from> <amount> <fee> <gas-price>",
		Short: "Bond token",
		Long:  "Bond token for useful activity",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			keyInfo, err := cliutil.GetLocalWalletInfo(valueFromFromFlag, kb, cdc, cliCtx)
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgBond("dummyAddress", addr, args[0], args[1], gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdUnbonding is the CLI command for unbonding
func GetCmdUnbonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond --from <from> <amount> <fee> <gas-price>",
		Short: "Unbond token",
		Long:  "Unbond token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			keyInfo, err := cliutil.GetLocalWalletInfo(valueFromFromFlag, kb, cdc, cliCtx)
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgUnBond("dummyAddress", addr, args[0], args[1], gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

// GetCmdCreateValidator implements the create validator command handler.
func GetCmdCreateValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-validator --from <from> --pubkey <validator_cons_pubkey> " +
			"[--moniker <moniker>] [--identity <identity>] [--website <site_address>] [--details <detail_description>]",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			keyInfo, err := cliutil.GetLocalWalletInfo(valueFromFromFlag, kb, cdc, cliCtx)
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())

			msg, err := BuildCreateValidatorMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FsPk)

	cmd.MarkFlagRequired(FlagPubKey)
	cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

// GetCmdEditValidator implements the create edit validator command.
func GetCmdEditValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "edit-validator --from <from> " +
			"[--moniker <moniker>] [--identity <identity>] [--website <site_address>] [--details <detail_description>]",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			valueFromFromFlag := viper.GetString(client.FlagFrom)
			keyInfo, err := cliutil.GetLocalWalletInfo(valueFromFromFlag, kb, cdc, cliCtx)
			if err != nil {
				return err
			}
			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())

			valAddr := cliCtx.GetFromAddress()
			description := types.Description{
				Moniker:  viper.GetString(FlagMoniker),
				Identity: viper.GetString(FlagIdentity),
				Website:  viper.GetString(FlagWebsite),
				Details:  viper.GetString(FlagDetails),
			}

			msg := types.NewMsgEditValidator(valAddr, description)

			// build and sign the transaction, then broadcast to Tendermint
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	cmd.Flags().AddFlagSet(fsDescriptionEdit)

	cmd.MarkFlagRequired(client.FlagFrom)

	return cmd
}

// BuildCreateValidatorMsg implements for adding validator module spec
func BuildCreateValidatorMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	valAddr := cliCtx.GetFromAddress()

	consPubKeyStr := viper.GetString(FlagPubKey)
	consPubKey, err := sdk.GetConsPubKeyBech32(consPubKeyStr)
	if err != nil {
		return types.MsgCreateValidator{}, err
	}

	description := types.NewDescription(
		viper.GetString(FlagMoniker),
		viper.GetString(FlagIdentity),
		viper.GetString(FlagWebsite),
		viper.GetString(FlagDetails),
	)

	msg := types.NewMsgCreateValidator(valAddr, consPubKey, description)

	return msg, nil
}
