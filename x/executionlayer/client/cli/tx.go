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
	"github.com/hdac-io/tendermint/crypto/secp256k1"

	"github.com/hdac-io/friday/client/context"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmdTransfer is the CLI command for transfer
func GetCmdTransfer(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf("transfer [--token-contract-address] "+
			"[--from] [--to-name or --to-pubkey or --to-%[1]spub] "+
			"[--amount] [--fee] [--gas-price]", sdk.Bech32MainPrefix),
		Short: "Transfer token",
		Long: fmt.Sprintf("Transfer token\n"+
			"It needs at least one of \"--to-name\", \"--to-pubkey\", or \"--to-%[1]spub\" parameter.\n"+
			"\t--to-name: readabld ID of recipient\n"+
			"\t--to-pubkey: Compressed Secp256k1 public keyof recipient\n"+
			"\t--to-%[1]spub: Bech32 encoded public key starting from '%[1]s' of recipient\n\n"+
			"\t--amount: amount to transfer\n"+
			"\t--fee: amount of fee\n"+
			"\t--gas-price: amount of gas price", sdk.Bech32MainPrefix),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// --token_contract_address: Currently useless
			contractAddress := viper.GetString(FlagTokenContractAddress)

			// --from: get key info from local wallet DB
			//         DO NOT GET PARAMETER AS ADDRESS, PUBKEY DIRECTLY. Only from the local-stored wallet
			// Get secp256k1 pubkey from local wallet key descriptor "--from"
			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			key, err := kb.Get(viper.GetString(client.FlagFrom))
			if err != nil {
				return err
			}

			fromPubkey := sdk.MustGetSecp256k1FromCryptoPubKey(key.GetPubKey())
			fromAddr := key.GetAddress()
			cliCtx = cliCtx.WithFromAddress(fromAddr)

			// --to-name, --to-pubkey, --to-[bech32_precix]pub
			// At least one of the above is essential
			// Derive secp256k1 pubkey from the value
			var toPubkey secp256k1.PubKeySecp256k1

			if toName := viper.GetString(FlagToName); toName != "" {
				// --to-name: from readable id
				toPubkeyPtr, err := cliutil.GetPubKey(cdc, cliCtx, toName)
				if err != nil {
					return err
				}
				toPubkey = *toPubkeyPtr
			} else if toRawPubkey := viper.GetString(FlagToPubkey); toRawPubkey != "" {
				// --to-pubkey: from raw secp256k1 public key
				toPubkeyPtr, err := sdk.GetSecp256k1FromRawHexString(toRawPubkey)
				if err != nil {
					return err
				}
				toPubkey = *toPubkeyPtr
			} else if toBech32Pubkey := viper.GetString(FlagToBech32Pubkey); toBech32Pubkey != "" {
				// --to-[bech32_prefix]pub: from bech32 public key (fridaypubxxxxxx...)
				rawToPubkey, err := sdk.GetSecp256k1FromBech32AccPubKey(toBech32Pubkey)
				if err != nil {
					return err
				}
				toPubkey = *rawToPubkey
			} else {
				return fmt.Errorf("at least one of --to-name, --to-pubkey, or --to-%[1]spub is essential", sdk.Bech32MainPrefix)
			}

			// get rest of the essential values
			amount, err := strconv.ParseUint(viper.GetString(FlagAmount), 10, 64)
			if err != nil {
				return err
			}
			fee, err := strconv.ParseUint(viper.GetString(FlagFee), 10, 64)
			if err != nil {
				return err
			}
			gasPrice, err := strconv.ParseUint(viper.GetString(FlagGasPrice), 10, 64)
			if err != nil {
				return err
			}

			// organize ABIs
			transferCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/transfer_to_account.wasm"))
			transferAbi := util.MakeArgsTransferToAccount(sdk.GetEEAddressFromSecp256k1PubKey(toPubkey).Bytes(), amount)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgTransfer(contractAddress, *fromPubkey, toPubkey, transferCode, transferAbi, paymentCode, paymentAbi, gasPrice, fromAddr)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "flag for custom local path of client's home dir")
	cmd.Flags().String(FlagTokenContractAddress, "", "flag for token contract address (Currently it is dummy)")
	cmd.Flags().String(client.FlagFrom, "", "flag for local wallet keybase loading")
	cmd.Flags().String(FlagToName, "", "flag for readable ID of recipient")
	cmd.Flags().String(FlagToPubkey, "", "flag for secp256k1 public key of receipeint")
	cmd.Flags().String(FlagToBech32Pubkey, "", "flag for bech32 public key of recipient")
	cmd.Flags().String(FlagAmount, "", "flag for the amount to transfer")
	cmd.Flags().String(FlagFee, "", "flag for tx fee")
	cmd.Flags().String(FlagGasPrice, "", "flag for gas price")

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagAmount)
	cmd.MarkFlagRequired(FlagFee)
	cmd.MarkFlagRequired(FlagGasPrice)

	return cmd
}

// GetCmdBonding is the CLI command for bonding
func GetCmdBonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond [--from] [--amount] [--fee] [--gas-price]",
		Short: "Bond token",
		Long: "Bond token\n" +
			"\t--from: alias of local-stored wallet\n" +
			"\t--amount: amount of bonding\n" +
			"\t--fee: amount of fee\n" +
			"\t--gas-price: amount of gas price",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// --from: get key info from local wallet DB
			//         DO NOT GET PARAMETER AS ADDRESS, PUBKEY DIRECTLY. Only from the local-stored wallet
			// Get secp256k1 pubkey from local wallet key descriptor "--from"
			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			key, err := kb.Get(viper.GetString(client.FlagFrom))
			if err != nil {
				return err
			}

			pubkey := sdk.MustGetSecp256k1FromCryptoPubKey(key.GetPubKey())
			addr := key.GetAddress()
			cliCtx = cliCtx.WithFromAddress(addr)

			amount := viper.GetUint64(FlagAmount)
			fee := viper.GetUint64(FlagFee)
			gasPrice := viper.GetUint64(FlagGasPrice)

			bondingCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/bonding.wasm"))
			bondingAbi := util.MakeArgsBonding(amount)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgBond(cliCtx.FromAddress.String(), *pubkey, bondingCode, bondingAbi, paymentCode, paymentAbi, gasPrice, addr)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "flag for custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Bech32 address")
	cmd.Flags().String(FlagFee, "", "fee")
	cmd.Flags().String(FlagGasPrice, "", "gas prices")
	cmd.Flags().AddFlagSet(FsAmount)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagFee)
	cmd.MarkFlagRequired(FlagGasPrice)
	cmd.MarkFlagRequired(FlagAmount)

	return cmd
}

// GetCmdUnbonding is the CLI command for unbonding
func GetCmdUnbonding(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond [--from] [--amount] [--fee] [--gas-price]",
		Short: "Unbond token",
		Long: "Unbond token\n" +
			"\t--from: alias of local-stored wallet\n" +
			"\t--amount: amount of unbonding\n" +
			"\t--fee: amount of fee\n" +
			"\t--gas-price: amount of gas price",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// --from: get key info from local wallet DB
			//         DO NOT GET PARAMETER AS ADDRESS, PUBKEY DIRECTLY. Only from the local-stored wallet
			// Get secp256k1 pubkey from local wallet key descriptor "--from"
			kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
			if err != nil {
				return err
			}

			key, err := kb.Get(viper.GetString(client.FlagFrom))
			if err != nil {
				return err
			}

			pubkey := sdk.MustGetSecp256k1FromCryptoPubKey(key.GetPubKey())
			addr := key.GetAddress()
			cliCtx = cliCtx.WithFromAddress(addr)

			amount := viper.GetUint64(FlagAmount)
			fee := viper.GetUint64(FlagFee)
			gasPrice := viper.GetUint64(FlagGasPrice)

			unbondingCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/unbonding.wasm"))
			unbondingAbi := util.MakeArgsUnBonding(amount)
			paymentCode := util.LoadWasmFile(os.ExpandEnv("$HOME/.nodef/contracts/standard_payment.wasm"))
			paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(fee))

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgUnBond(cliCtx.FromAddress.String(), *pubkey, unbondingCode, unbondingAbi, paymentCode, paymentAbi, gasPrice, addr)
			txBldr = txBldr.WithGas(gasPrice)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "flag for custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Bech32 address")
	cmd.Flags().String(FlagFee, "", "fee")
	cmd.Flags().String(FlagGasPrice, "", "gas prices")
	cmd.Flags().AddFlagSet(FsAmount)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagFee)
	cmd.MarkFlagRequired(FlagGasPrice)
	cmd.MarkFlagRequired(FlagAmount)

	return cmd
}

// GetCmdCreateValidator implements the create validator command handler.
func GetCmdCreateValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator [--from] [--pubkey] [--moniker] [--identity] [--website] [--details]",
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

	cmd.Flags().String(client.FlagHome, DefaultClientHome, "flag for custom local path of client's home dir")
	cmd.Flags().String(client.FlagFrom, "", "Bech32 address")
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(FsPk)

	cmd.MarkFlagRequired(client.FlagFrom)
	cmd.MarkFlagRequired(FlagPubKey)
	cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

func BuildCreateValidatorMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	kb, err := client.NewKeyBaseFromDir(viper.GetString(client.FlagHome))
	if err != nil {
		return types.MsgCreateValidator{}, err
	}

	key, err := kb.Get(viper.GetString(client.FlagFrom))
	valPubKey := sdk.MustGetSecp256k1FromCryptoPubKey(key.GetPubKey())
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
