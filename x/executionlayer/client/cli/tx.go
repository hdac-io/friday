package cli

import (
	"fmt"
	"math/big"
	"os"
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
		Use:   "run <type> <wasm-path>|<uref>|<name>|<hash> <argument> <fee> <gas_price> --wallet|--address|--nickname <from>",
		Short: "Run contract",
		Long: "Run contract\n" +
			"There are 4 types of contract run. ('wasm', 'uref', 'name', 'hash)",
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Extract "from" from flags
			var fromAddr sdk.AccAddress

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			if walletname := viper.GetString(FlagWallet); walletname != "" {
				key, err := kb.Get(walletname)
				if err != nil {
					return err
				}

				fromAddr = key.GetAddress()
				cliCtx = cliCtx.WithFromAddress(fromAddr).WithFromName(key.GetName())
			} else if straddr := viper.GetString(FlagAddress); straddr != "" {
				fromAddr, err = sdk.AccAddressFromBech32(straddr)
				if err != nil {
					return fmt.Errorf("malformed address in --address: %s\n%s", straddr, err.Error())
				}

				key, err := kb.GetByAddress(fromAddr)
				if err != nil {
					return err
				}

				cliCtx = cliCtx.WithFromAddress(fromAddr).WithFromName(key.GetName())
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				fromAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("no registered address of the given nickname '%s'", nickname)
				}

				key, err := kb.GetByAddress(fromAddr)
				if err != nil {
					return err
				}
				cliCtx = cliCtx.WithFromAddress(fromAddr).WithFromName(key.GetName())
			} else {
				return fmt.Errorf("one of --address, --wallet, --nickname is essential")
			}

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

			sessionArgs, err := cliutil.ProtobufSafeDecodeingHexString(args[2])
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

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgExecute(
				"dummyAddress",
				fromAddr,
				sessionType,
				sessionCode,
				sessionArgs,
				util.WASM,
				util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm")),
				util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee)),
				gasPrice,
			)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "Bech32 endocded address (fridayxxxxxx..)")
	cmd.Flags().String(FlagWallet, "", "Wallet alias")
	cmd.Flags().String(FlagNickname, "", "Nickname")

	return cmd
}

// GetCmdTransfer is the CLI command for transfer
func GetCmdTransfer(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-to <recipient_nickname>|<address> <amount> <fee> <gas_price> --wallet|--address|--nickname <from>",
		Short: "Transfer Hdac token",
		Long: "Transfer Hdac token\n" +
			"It needs at least one of '--wallet', '--address', or '--nickname' flag.",
		Args: cobra.ExactArgs(4),
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
			amount, err := strconv.ParseUint(args[1], 10, 64)
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

			// Extract "from" from flags
			var fromAddr sdk.AccAddress

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			if walletname := viper.GetString(FlagWallet); walletname != "" {
				key, err := kb.Get(walletname)
				if err != nil {
					return err
				}

				fromAddr = key.GetAddress()
				cliCtx = cliCtx.WithFromAddress(fromAddr).WithFromName(key.GetName())
			} else if straddr := viper.GetString(FlagAddress); straddr != "" {
				fromAddr, err = sdk.AccAddressFromBech32(straddr)
				if err != nil {
					return fmt.Errorf("malformed address in --address: %s\n%s", straddr, err.Error())
				}

				key, err := kb.GetByAddress(fromAddr)
				if err != nil {
					return err
				}

				cliCtx = cliCtx.WithFromAddress(fromAddr).WithFromName(key.GetName())
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				fromAddr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("no registered address of the given nickname '%s'", nickname)
				}

				key, err := kb.GetByAddress(fromAddr)
				if err != nil {
					return err
				}
				cliCtx = cliCtx.WithFromAddress(fromAddr).WithFromName(key.GetName())
			} else {
				return fmt.Errorf("one of --address, --wallet, --nickname is essential")
			}

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgTransfer("dummyAddress", fromAddr, recipentAddr, amount, fee, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "Bech32 endocded address (fridayxxxxxx..)")
	cmd.Flags().String(FlagWallet, "", "Wallet alias")
	cmd.Flags().String(FlagNickname, "", "Nickname")

	return cmd
}

// GetCmdBonding is the CLI command for bonding
func GetCmdBonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond --wallet|--address|--nickname <from> <amount> <fee> <gas-price>",
		Short: "Bond token",
		Long:  "Bond token for useful activity",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var addr sdk.AccAddress
			var err error

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			// Extract "from" from flags
			if walletname := viper.GetString(FlagWallet); walletname != "" {
				key, err := kb.Get(walletname)
				if err != nil {
					return err
				}

				addr = key.GetAddress()
				cliCtx = cliCtx.WithFromAddress(addr).WithFromName(key.GetName())
			} else if straddr := viper.GetString(FlagAddress); straddr != "" {
				addr, err = sdk.AccAddressFromBech32(straddr)
				if err != nil {
					return fmt.Errorf("malformed address in --address: %s\n%s", straddr, err.Error())
				}

				key, err := kb.GetByAddress(addr)
				if err != nil {
					return err
				}
				cliCtx = cliCtx.WithFromAddress(addr).WithFromName(key.GetName())
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("no registered address of the given nickname '%s'", nickname)
				}
				key, err := kb.GetByAddress(addr)
				if err != nil {
					return err
				}
				cliCtx = cliCtx.WithFromAddress(addr).WithFromName(key.GetName())
			} else {
				return fmt.Errorf("one of --address, --wallet, --nickname is essential")
			}

			// Numbers parsing
			amount, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			fee, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			gasPrice, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgBond("dummyAddress", addr, amount, fee, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "Bech32 endocded address (fridayxxxxxx..)")
	cmd.Flags().String(FlagWallet, "", "Wallet alias")
	cmd.Flags().String(FlagNickname, "", "Nickname")

	return cmd
}

// GetCmdUnbonding is the CLI command for unbonding
func GetCmdUnbonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond --wallet|--address|--nickname <from> <amount> <fee> <gas-price>",
		Short: "Unbond token",
		Long:  "Unbond token for converts tokens as a freedom",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var addr sdk.AccAddress
			var err error

			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			// Extract "from" from flags
			if walletname := viper.GetString(FlagWallet); walletname != "" {
				key, err := kb.Get(walletname)
				if err != nil {
					return err
				}

				addr = key.GetAddress()
				cliCtx = cliCtx.WithFromAddress(addr).WithFromName(key.GetName())
			} else if straddr := viper.GetString(FlagAddress); straddr != "" {
				addr, err = sdk.AccAddressFromBech32(straddr)
				if err != nil {
					return fmt.Errorf("malformed address in --address: %s\n%s", straddr, err.Error())
				}

				key, err := kb.GetByAddress(addr)
				if err != nil {
					return err
				}
				cliCtx = cliCtx.WithFromAddress(addr).WithFromName(key.GetName())
			} else if nickname := viper.GetString(FlagNickname); nickname != "" {
				addr, err = cliutil.GetAddress(cliCtx.Codec, cliCtx, nickname)
				if err != nil {
					return fmt.Errorf("no registered address of the given nickname '%s'", nickname)
				}

				key, err := kb.GetByAddress(addr)
				if err != nil {
					return err
				}
				cliCtx = cliCtx.WithFromAddress(addr).WithFromName(key.GetName())
			} else {
				return fmt.Errorf("one of --address, --wallet, --nickname is essential")
			}

			// Numbers parsing
			amount, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			fee, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			gasPrice, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithFromAddress(addr)

			// build and sign the transaction, then broadcast to Tendermint
			// TODO: Currently implementation of contract address is dummy
			msg := types.NewMsgUnBond("dummyAddress", addr, amount, fee, gasPrice)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(FlagAddress, "", "Bech32 endocded address (fridayxxxxxx..)")
	cmd.Flags().String(FlagWallet, "", "Wallet alias")
	cmd.Flags().String(FlagNickname, "", "Nickname")

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

			msg, err := BuildCreateValidatorMsg(cliCtx)
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "Custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Bech32 encoded address (fridayxxxxxx...)")
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FsPk)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagPubKey)
	cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

// BuildCreateValidatorMsg implements for adding validator module spec
func BuildCreateValidatorMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
	if err != nil {
		return types.MsgCreateValidator{}, err
	}

	key, err := kb.Get(viper.GetString(client.FlagFrom))
	valPubKey := key.GetPubKey()
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

	msg := types.NewMsgCreateValidator(sdk.AccAddress(valAddr), valPubKey, consPubKey, description)

	return msg, nil
}
