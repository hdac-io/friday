package cli

import (
	"github.com/spf13/cobra"
	"math/big"
	"os"
	"regexp"
	"strconv"

	"github.com/hdac-io/friday/client"
	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/hdac-io/friday/x/bank/internal/types"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	ee "github.com/hdac-io/friday/x/executionlayer/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		SendTxCmd(cdc),
	)
	return txCmd
}

// SendTxCmd will create a send tx and sign it with the given key.
func SendTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount] [fee] [gas_price]",
		Short: "Create and sign a send tx",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithFrom(args[0]).WithCodec(cdc)

			to, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			// parse coins trying to be sent
			coins, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgSend(cliCtx.GetFromAddress(), to, coins)
			eeMsg := makeEETransferMsg(cdc, cliCtx, to, args[2], args[3], gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg, eeMsg})
		},
	}

	cmd = client.PostCommands(cmd)[0]

	return cmd
}

func makeEETransferMsg(cdc *codec.Codec, cliCtx context.CLIContext, to sdk.AccAddress, strAmount string, strFee string, gasPrice uint64) sdk.Msg {
	toPublicKey := ee.ToPublicKey(to)

	re := regexp.MustCompile("[0-9]+")
	coins, err := strconv.ParseUint(re.FindAllString(strAmount, -1)[0], 10, 64)
	if err != nil {
		return ee.MsgExecute{}
	}

	fee, err := strconv.ParseUint(strFee, 10, 64)
	if err != nil {
		return ee.MsgExecute{}
	}

	transferCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
	transferAbi := util.MakeArgsTransferToAccount(toPublicKey, coins)
	paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
	paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

	// build and sign the transaction, then broadcast to Tendermint
	msg := ee.NewMsgExecute([]byte{0}, cliCtx.FromAddress, cliCtx.FromAddress, transferCode, transferAbi, paymentCode, paymentAbi, gasPrice)
	return msg
}
