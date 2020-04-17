package types

import (
	"fmt"
	"strings"

	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// nolint
const (
	// TODO: Why can't we just have one string description which can be JSON by convention
	MaxMonikerLength  = 70
	MaxIdentityLength = 3000
	MaxWebsiteLength  = 140
	MaxDetailsLength  = 280
)

// Validator - save a validater information
type Validator struct {
	OperatorAddress sdk.AccAddress `json:"operator_address" yaml:"operator_address"` // address of the validator's operator; bech encoded in JSON
	ConsPubKey      crypto.PubKey  `json:"consensus_pubkey" yaml:"consensus_pubkey"` // the consensus public key of the validator; bech encoded in JSON
	Description     Description    `json:"description" yaml:"description"`           // description terms for the validator
	Stake           string         `json:"stake" yaml:"stake"`
}

// NewValidator - initialize a new validator
func NewValidator(operator sdk.AccAddress, pubKey crypto.PubKey, description Description, stake string) Validator {
	return Validator{
		OperatorAddress: operator,
		ConsPubKey:      pubKey,
		Description:     description,
		Stake:           stake,
	}
}

// return the redelegation
func MustMarshalValidator(cdc *codec.Codec, validator Validator) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(validator)
}

// unmarshal a redelegation from a store value
func MustUnmarshalValidator(cdc *codec.Codec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// unmarshal a redelegation from a store value
func UnmarshalValidator(cdc *codec.Codec, value []byte) (validator Validator, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &validator)
	return validator, err
}

// String returns a human readable string representation of a validator.
func (v Validator) String() string {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`Validator
  Operator Address:           %s
  Validator Consensus Pubkey: %s
  Description:                %s
  Stake:					  %s`, v.OperatorAddress, bechConsPubKey, v.Description, v.Stake)
}

// constant used in flags to indicate that description field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker" yaml:"moniker"`   // name
	Identity string `json:"identity" yaml:"identity"` // optional identity signature (ex. UPort or Keybase)
	Website  string `json:"website" yaml:"website"`   // optional website link
	Details  string `json:"details" yaml:"details"`   // optional details
}

// NewDescription returns a new Description with the provided values.
func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// UpdateDescription updates the fields of a given description. An error is
// returned if the resulting description contains an invalid length.
func (d Description) UpdateDescription(d2 Description) (Description, sdk.Error) {
	if d2.Moniker == DoNotModifyDesc {
		d2.Moniker = d.Moniker
	}
	if d2.Identity == DoNotModifyDesc {
		d2.Identity = d.Identity
	}
	if d2.Website == DoNotModifyDesc {
		d2.Website = d.Website
	}
	if d2.Details == DoNotModifyDesc {
		d2.Details = d.Details
	}

	return Description{
		Moniker:  d2.Moniker,
		Identity: d2.Identity,
		Website:  d2.Website,
		Details:  d2.Details,
	}.EnsureLength()
}

// EnsureLength ensures the length of a validator's description.
func (d Description) EnsureLength() (Description, sdk.Error) {
	if len(d.Moniker) > MaxMonikerLength {
		return d, ErrDescriptionLength(DefaultCodespace, "moniker", len(d.Moniker), MaxMonikerLength)
	}
	if len(d.Identity) > MaxIdentityLength {
		return d, ErrDescriptionLength(DefaultCodespace, "identity", len(d.Identity), MaxIdentityLength)
	}
	if len(d.Website) > MaxWebsiteLength {
		return d, ErrDescriptionLength(DefaultCodespace, "website", len(d.Website), MaxWebsiteLength)
	}
	if len(d.Details) > MaxDetailsLength {
		return d, ErrDescriptionLength(DefaultCodespace, "details", len(d.Details), MaxDetailsLength)
	}

	return d, nil
}

// this is a helper struct used for JSON de- and encoding only
type bechValidator struct {
	Address     string      `json:"address", yaml:"address"`
	ConsPubKey  string      `json:"consensus_pubkey" yaml:"consensus_pubkey"` // the bech32 consensus public key of the validator
	Description Description `json:"description" yaml:"description"`           // description terms for the validator
	Stake       string      `json:"stake" yaml:"stake"`
}

// MarshalJSON marshals the validator to JSON using Bech32
func (v Validator) MarshalJSON() ([]byte, error) {
	bechConsPubKey, err := sdk.Bech32ifyConsPub(v.ConsPubKey)
	if err != nil {
		return nil, err
	}

	return codec.Cdc.MarshalJSON(bechValidator{
		Address:     v.OperatorAddress.String(),
		ConsPubKey:  bechConsPubKey,
		Description: v.Description,
		Stake:       v.Stake,
	})
}

// UnmarshalJSON unmarshals the validator from JSON using Bech32
func (v *Validator) UnmarshalJSON(data []byte) error {
	bv := &bechValidator{}
	if err := codec.Cdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	valAddress, err := sdk.AccAddressFromBech32(bv.Address)
	if err != nil {
		return err
	}
	consPubKey, err := sdk.GetConsPubKeyBech32(bv.ConsPubKey)
	if err != nil {
		return err
	}
	*v = Validator{
		OperatorAddress: valAddress,
		ConsPubKey:      consPubKey,
		Description:     bv.Description,
		Stake:           bv.Stake,
	}
	return nil
}

// only the vitals
func (v Validator) TestEquivalent(v2 Validator) bool {
	return v.ConsPubKey.Equals(v2.ConsPubKey) &&
		v.Description == v2.Description &&
		v.Stake == v2.Stake
}

// return the TM validator address
func (v Validator) ConsAddress() sdk.ConsAddress {
	return sdk.ConsAddress(v.ConsPubKey.Address())
}

// Validators is a collection of Validator
type Validators []Validator

func (v Validators) String() (out string) {
	for _, val := range v {
		out += val.String() + "\n"
	}
	return strings.TrimSpace(out)
}

type Delegator struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Amount  string         `json:"amount" yaml:"amount"`
}

// NewDelegator - initialize a new delegator
func NewDelegator(address sdk.AccAddress, amount string) Delegator {
	return Delegator{
		Address: address,
		Amount:  amount,
	}
}

// return the delegate
func MustMarshalDelegator(cdc *codec.Codec, delegator Delegator) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(delegator)
}

// unmarshal a delegator from a store value
func MustUnmarshalDelegator(cdc *codec.Codec, value []byte) Delegator {
	delegator, err := UnmarshalDelegator(cdc, value)
	if err != nil {
		panic(err)
	}
	return delegator
}

// unmarshal a delegator from a store value
func UnmarshalDelegator(cdc *codec.Codec, value []byte) (delegator Delegator, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &delegator)
	return delegator, err
}

// String returns a human readable string representation of a validator.
func (d Delegator) String() string {
	return fmt.Sprintf(`Deligator
  Address:           %s
  Amount:			 %s`, d.Address, d.Amount)
}

// Delegators is a collection of Delegator
type Delegators []Delegator

func (d Delegators) String() (out string) {
	for _, val := range d {
		out += val.String() + "\n"
	}
	return strings.TrimSpace(out)
}

type Voter struct {
	Address []byte `json:"address" yaml:"address"`
	Amount  string `json:"amount" yaml:"amount"`
}

// NewVoter - initialize a new voter
func NewVoter(address sdk.EEAddress, amount string) Voter {
	return Voter{
		Address: address,
		Amount:  amount,
	}
}

// return the voter
func MustMarshalVoter(cdc *codec.Codec, voter Voter) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(voter)
}

// unmarshal a delegator from a store value
func MustUnmarshalVoter(cdc *codec.Codec, value []byte) Voter {
	voter, err := UnmarshalVoter(cdc, value)
	if err != nil {
		panic(err)
	}
	return voter
}

// unmarshal a voter from a store value
func UnmarshalVoter(cdc *codec.Codec, value []byte) (voter Voter, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &voter)
	return voter, err
}

// String returns a human readable string representation of a validator.
func (d Voter) String() string {
	return fmt.Sprintf(`Voters
  Address:           %s
  Amount:			 %s`, d.Address, d.Amount)
}

// Voters is a collection of Delegator
type Voters []Voter

func (d Voters) String() (out string) {
	for _, val := range d {
		out += val.String() + "\n"
	}
	return strings.TrimSpace(out)
}
