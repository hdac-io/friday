package cli

import (
	"fmt"
	"strconv"
	"strings"

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
			var contractAddress string

			switch sessionType {
			case util.WASM:
				contractAddress = "wasm_file_direct_execution"
				sessionCode = util.LoadWasmFile(args[1])
			case util.HASH:
				contractAddress = args[1]
				contractHashAddr, err := sdk.ContractHashAddressFromBech32(args[1])
				if err != nil {
					return err
				}
				sessionCode = contractHashAddr.Bytes()
			case util.UREF:
				contractAddress = args[1]
				contractUrefAddr, err := sdk.ContractUrefAddressFromBech32(args[1])
				if err != nil {
					return err
				}
				sessionCode = contractUrefAddr.Bytes()
			case util.NAME:
				contractAddress = fmt.Sprintf("%s:%s", fromAddr.String(), args[1])
				sessionCode = []byte(args[1])
			default:
				return fmt.Errorf("type must be one of wasm, name, uref, or hash")
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[3]))
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgExecute(
				contractAddress,
				fromAddr,
				sessionType,
				sessionCode,
				args[2],
				string(fee),
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

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[2]))
			if err != nil {
				return err
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
			msg := types.NewMsgTransfer("system:transfer", fromAddr, recipentAddr, string(amount), string(fee), gasPrice)
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

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[0]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
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
			msg := types.NewMsgBond("system:bond", addr, string(amount), string(fee), gasPrice)
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

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[0]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
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
			msg := types.NewMsgUnBond("system:unbond", addr, string(amount), string(fee), gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

func GetCmdDelegate(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate <validator-address> <amount> <fee> <gas-price> --from <from>",
		Short: "Delegate token",
		Long:  "Delegate token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(4),
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

			valAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				valAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, args[0])
				if err != nil {
					return fmt.Errorf("no nickname mapping of %s", args[0])
				}
			}

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[2]))
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgDelegate("system:delegate", addr, valAddress, string(amount), string(fee), gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

func GetCmdUndelegate(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate <validator-address> <amount> <fee> <gas-price> --from <from>",
		Short: "Undelegate token",
		Long:  "Undelegate token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(4),
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

			valAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				valAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, args[0])
				if err != nil {
					return fmt.Errorf("no nickname mapping of %s", args[0])
				}
			}

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[2]))
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgUndelegate("system:undelegate", addr, valAddress, string(amount), string(fee), gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

func GetCmdRedelegate(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate <src-validator-address> <dest-validator-address> <amount> <fee> <gas-price> --from <from>",
		Short: "Redelegate token",
		Long:  "Redelegate token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(5),
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

			srcValAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				srcValAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, args[0])
				if err != nil {
					return fmt.Errorf("no nickname mapping of %s", args[0])
				}
			}

			destValAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				destValAddress, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, args[0])
				if err != nil {
					return fmt.Errorf("no nickname mapping of %s", args[0])
				}
			}

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[2]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[3]))
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgRedelegate("system:redelegate", addr, srcValAddress, destValAddress, string(amount), string(fee), gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

func GetCmdVote(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote <contract_address> <amount> <fee> <gas-price> --from <from>",
		Short: "Vote token",
		Long:  "Vote token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(4),
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

			var contractAddress sdk.ContractAddress
			if strings.HasPrefix(args[0], sdk.Bech32PrefixContractURef) {
				contractAddress, err = sdk.ContractUrefAddressFromBech32(args[0])
			} else if strings.HasPrefix(args[0], sdk.Bech32PrefixContractHash) {
				contractAddress, err = sdk.ContractHashAddressFromBech32(args[0])
			} else {
				err = fmt.Errorf("Malformed contract address")
			}

			if err != nil {
				return err
			}

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[2]))
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgVote("system:vote", addr, contractAddress, string(amount), string(fee), gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

func GetCmdUnvote(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unvote <contract_address> <amount> <fee> <gas-price> --from <from>",
		Short: "Unvote token",
		Long:  "Unvote token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(4),
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

			var contractAddress sdk.ContractAddress
			if strings.HasPrefix(args[0], sdk.Bech32PrefixContractURef) {
				contractAddress, err = sdk.ContractUrefAddressFromBech32(args[0])
			} else if strings.HasPrefix(args[0], sdk.Bech32PrefixContractHash) {
				contractAddress, err = sdk.ContractHashAddressFromBech32(args[0])
			} else {
				return fmt.Errorf("Malformed contract address")
			}

			amount, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
			if err != nil {
				return err
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[2]))
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(keyInfo.GetAddress()).WithFromName(keyInfo.GetName())
			addr := keyInfo.GetAddress()

			gasPrice, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgUnvote("system:unvote", addr, contractAddress, string(amount), string(fee), gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Executor's identity (one of wallet alias, address, nickname)")

	return cmd
}

func GetCmdClaimReward(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim reward|commission <fee> <gas-price> --from <from>",
		Short: "Reward or commission token",
		Long:  "Reward for delegated quantity",
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

			var isRewardOrCommission bool
			switch args[0] {
			case types.RewardString:
				isRewardOrCommission = types.RewardValue
			case types.CommissionString:
				isRewardOrCommission = types.CommissionValue
			}

			fee, err := cliutil.ToBigsun(cliutil.Hdac(args[1]))
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgClaim(fmt.Sprintf("system:claim_%s", args[0]), addr, isRewardOrCommission, string(fee), gasPrice)
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
