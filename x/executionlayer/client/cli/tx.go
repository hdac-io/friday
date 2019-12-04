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
		GetCmdTransfer(cdc), GetCmdBonding(cdc), GetCmdUnbonding(cdc),
	)...)
	return executionlayerTxCmd
}

// The code below is a pattern of sending Tx
// You may start from the example first

// GetCmdTransfer is the CLI command for transfer
func GetCmdTransfer(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [from_key_or_address] [to_address] [amount] [fee] [gas_price]",
		Short: "Create and sign a send tx",
		Args:  cobra.ExactArgs(5), // # of arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			to, err := sdk.AccAddressFromBech32(args[1])
			toPublicKey := types.ToPublicKey(to)

			if err != nil {
				return err
			}

			coins, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			fee, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			transferCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
			transferAbi := util.MakeArgsTransferToAccount(toPublicKey, coins)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgExecute([]byte{0}, cliCtx.FromAddress, cliCtx.FromAddress, transferCode, transferAbi, paymentCode, paymentAbi, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdBonding is the CLI command for bonding
func GetCmdBonding(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "bond [from_key_or_address] [amount] [fee] [gas_price]",
		Short: "Create and sign a bonding tx",
		Args:  cobra.ExactArgs(4), // # of arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			coins, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			fee, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			bondingCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/bonding.wasm"))
			bondingAbi := util.MakeArgsBonding(coins)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgExecute([]byte{0}, cliCtx.FromAddress, cliCtx.FromAddress, bondingCode, bondingAbi, paymentCode, paymentAbi, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdUnbonding is the CLI command for unbonding
func GetCmdUnbonding(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbond [from_key_or_address] [amount] [fee] [gas_price]",
		Short: "Create and sign a unbonding tx",
		Args:  cobra.ExactArgs(4), // # of arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			coins, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			fee, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			unbondingCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/unbonding.wasm"))
			unbondingAbi := util.MakeArgsBonding(coins)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgExecute([]byte{0}, cliCtx.FromAddress, cliCtx.FromAddress, unbondingCode, unbondingAbi, paymentCode, paymentAbi, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
