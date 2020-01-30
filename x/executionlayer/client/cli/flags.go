package cli

import (
	"os"

	flag "github.com/spf13/pflag"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// nolint
const (
	FlagAddressValidator     = "validator"
	FlagAddressValidatorSrc  = "addr-validator-source"
	FlagAddressValidatorDst  = "addr-validator-dest"
	FlagPubKey               = "pubkey"
	FlagTokenContractAddress = "token-contract-address"
	FlagBech32PubKey         = sdk.Bech32MainPrefix + "pub"
	FlagAmount               = "amount"
	FlagFee                  = "fee"
	FlagGasPrice             = "gas-price"
	FlagBlockHash            = "blockhash"
	FlagConsPubKey           = "cons-" + FlagPubKey
	FlagConsBech32PubKey     = "cons-" + FlagBech32PubKey
	FlagValPubkey            = "val-" + FlagPubKey
	FlagValBech32PubKey      = "val-" + FlagBech32PubKey

	FlagWallet   = "wallet"
	FlagAddress  = "address"
	FlagNickname = "nickname"

	FlagMoniker  = "moniker"
	FlagIdentity = "identity"
	FlagWebsite  = "website"
	FlagDetails  = "details"

	FlagMinSelfDelegation = "min-self-delegation"

	FlagGenesisFormat = "genesis-format"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"
)

// common flagsets to add to various functions
var (
	FsPk                = flag.NewFlagSet("", flag.ContinueOnError)
	FsAmount            = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionCreate = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionEdit   = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator         = flag.NewFlagSet("", flag.ContinueOnError)

	DefaultClientHome = os.ExpandEnv("$HOME/.clif")
)

func init() {
	FsPk.String(FlagPubKey, "", "The Bech32 encoded PubKey of the validator")
	FsAmount.String(FlagAmount, "", "Amount of coins to bond")
	fsDescriptionCreate.String(FlagMoniker, "", "The validator's name")
	fsDescriptionCreate.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fsDescriptionCreate.String(FlagWebsite, "", "The validator's (optional) website")
	fsDescriptionCreate.String(FlagDetails, "", "The validator's (optional) details")
	fsDescriptionEdit.String(FlagMoniker, types.DoNotModifyDesc, "The validator's name")
	fsDescriptionEdit.String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	fsDescriptionEdit.String(FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	fsDescriptionEdit.String(FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")
	fsValidator.String(FlagAddressValidator, "", "The Bech32 address of the validator")
}
