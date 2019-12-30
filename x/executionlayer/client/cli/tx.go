package cli

import (
	"math/big"
	"os"
	"strconv"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/executionlayer/types"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmdTransfer is the CLI command for transfer
func GetCmdTransfer(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [token_contract_address] [from_address] [to_address] [amount] [fee] [gas_price]",
		Short: "Transfer token",
		Args:  cobra.ExactArgs(6), // # of arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[1]).WithCodec(cdc)

			tokenOwnerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			fromAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			toAddress, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			toPublicKey := types.ToPublicKey(toAddress)
			amount, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}
			fee, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}
			gasPrice, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return err
			}

			transferCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
			transferAbi := util.MakeArgsTransferToAccount(toPublicKey, amount)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgTransfer(tokenOwnerAddress, fromAddress, toAddress, transferCode, transferAbi, paymentCode, paymentAbi, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdBonding is the CLI command for bonding
func GetCmdBonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond",
		Short: "Create and sign a bonding tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddress, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			amount := viper.GetUint64(FlagAmount)

			fee := viper.GetUint64(FlagFee)

			gasPrice := viper.GetUint64(FlagGasPrice)

			bondingCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/bonding.wasm"))
			bondingAbi := util.MakeArgsBonding(amount)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgBond(cliCtx.FromAddress, valAddress, bondingCode, bondingAbi, paymentCode, paymentAbi, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(client.FlagFrom, "", "The Bech32 address")
	cmd.Flags().String(FlagFee, "", "fee")
	cmd.Flags().String(FlagGasPrice, "", "gas prices")
	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(FsAmount)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagFee)
	cmd.MarkFlagRequired(FlagGasPrice)
	cmd.MarkFlagRequired(FlagAddressValidator)
	cmd.MarkFlagRequired(FlagAmount)

	return cmd
}

// GetCmdUnbonding is the CLI command for unbonding
func GetCmdUnbonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "Create and sign a unbonding tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddress, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			amount := viper.GetUint64(FlagAmount)

			fee := viper.GetUint64(FlagFee)

			gasPrice := viper.GetUint64(FlagGasPrice)

			unbondingCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/unbonding.wasm"))
			unbondingAbi := util.MakeArgsUnBonding(amount)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgUnBond(cliCtx.FromAddress, valAddress, unbondingCode, unbondingAbi, paymentCode, paymentAbi, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagFrom, "", "The Bech32 address")
	cmd.Flags().String(FlagFee, "", "fee")
	cmd.Flags().String(FlagGasPrice, "", "gas prices")
	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(FsAmount)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagFee)
	cmd.MarkFlagRequired(FlagGasPrice)
	cmd.MarkFlagRequired(FlagAddressValidator)
	cmd.MarkFlagRequired(FlagAmount)

	return cmd
}

// GetCmdCreateValidator implements the create validator command handler.
func GetCmdCreateValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			msg, err := BuildCreateValidatorMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagFrom, "", "from address")
	cmd.Flags().AddFlagSet(FsPk)
	cmd.Flags().AddFlagSet(fsDescriptionCreate)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagPubKey)
	cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

func BuildCreateValidatorMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	valAddr := cliCtx.GetFromAddress()
	pkStr := viper.GetString(FlagPubKey)

	pk, err := sdk.GetConsPubKeyBech32(pkStr)
	if err != nil {
		return types.MsgCreateValidator{}, err
	}

	description := types.NewDescription(
		viper.GetString(FlagMoniker),
		viper.GetString(FlagIdentity),
		viper.GetString(FlagWebsite),
		viper.GetString(FlagDetails),
	)

	msg := types.NewMsgCreateValidator(sdk.ValAddress(valAddr), pk, description)

	return msg, nil
}
