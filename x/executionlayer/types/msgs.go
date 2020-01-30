package types

import (
	"bytes"
	"encoding/json"

	"github.com/hdac-io/tendermint/crypto"
	secp256k1 "github.com/hdac-io/tendermint/crypto/secp256k1"

	"github.com/hdac-io/friday/types"
	sdk "github.com/hdac-io/friday/types"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

// MsgExecute for sending deploy to execution engine
type MsgExecute struct {
	ContractAddress string                    `json:"contract_address"`
	ExecPubkey      secp256k1.PubKeySecp256k1 `json:"exec_pubkey"`
	SessionCode     []byte                    `json:"session_code"`
	SessionArgs     []byte                    `json:"session_args"`
	PaymentCode     []byte                    `json:"payment_code"`
	PaymentArgs     []byte                    `json:"payment_args"`
	GasPrice        uint64                    `json:"gas_price"`
	Signer          sdk.AccAddress            `json:"signer"`
}

// NewMsgExecute is a constructor function for MsgSetName
func NewMsgExecute(
	contractAddress string,
	execPubkey secp256k1.PubKeySecp256k1,
	sessionCode, sessionArgs []byte,
	paymentCode, paymentArgs []byte,
	gasPrice uint64,
	signer sdk.AccAddress,
) MsgExecute {
	return MsgExecute{
		ExecPubkey:      execPubkey,
		ContractAddress: contractAddress,
		SessionCode:     sessionCode,
		SessionArgs:     sessionArgs,
		PaymentCode:     paymentCode,
		PaymentArgs:     paymentArgs,
		GasPrice:        gasPrice,
		Signer:          signer,
	}
}

// Route should return the name of the module
func (msg MsgExecute) Route() string { return RouterKey }

// Type should return the action
func (msg MsgExecute) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgExecute) ValidateBasic() sdk.Error {
	if msg.Signer.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgExecute) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgExecute) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// MsgTransfer for sending deploy to execution engine
type MsgTransfer struct {
	ToPubkey   secp256k1.PubKeySecp256k1 `json:"to_pubkey"`
	MsgExecute MsgExecute                `json:"msg_execute"`
}

// NewMsgTransfer is a constructor function for MsgSetName
func NewMsgTransfer(
	tokenContractAddress string,
	fromPubkey, toPubkey secp256k1.PubKeySecp256k1,
	transferCode, transferArgs, paymentCode, paymentArgs []byte,
	gasPrice uint64,
	signer sdk.AccAddress,
) MsgTransfer {
	return MsgTransfer{
		ToPubkey: toPubkey,
		MsgExecute: MsgExecute{
			ContractAddress: tokenContractAddress,
			ExecPubkey:      fromPubkey,
			SessionCode:     transferCode,
			SessionArgs:     transferArgs,
			PaymentCode:     paymentCode,
			PaymentArgs:     paymentArgs,
			GasPrice:        gasPrice,
			Signer:          signer,
		},
	}
}

// Route should return the name of the module
func (msg MsgTransfer) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTransfer) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if msg.MsgExecute.Signer.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.MsgExecute.Signer}
}

//______________________________________________________________________
// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	ValidatorAddress sdk.AccAddress `json:"validator_address" yaml:"validator_address"`
	ValidatorPubKey  crypto.PubKey  `json:"validator_pubkey" yaml:"validator_pubkey"`
	ConsPubKey       crypto.PubKey  `json:"cons_pubkey" yaml:"cons_pubkey"`
	Description      Description    `json:"description" yaml:"description"`
}

type msgCreateValidatorJSON struct {
	ValidatorAddress sdk.AccAddress `json:"validator_address" yaml:"validator_address"`
	ValidatorPubKey  string         `json:"validator_pubkey" yaml:"validator_pubkey"`
	ConsPubKey       string         `json:"cons_pubkey" yaml:"cons_pubkey"`
	Description      Description    `json:"description" yaml:"description"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddress sdk.AccAddress,
	valPubKey crypto.PubKey,
	consPubKey crypto.PubKey,
	description Description,
) MsgCreateValidator {
	return MsgCreateValidator{
		ValidatorAddress: valAddress,
		ValidatorPubKey:  valPubKey,
		ConsPubKey:       consPubKey,
		Description:      description,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	_, err := types.GetEEAddressFromCryptoPubkey(msg.ValidatorPubKey)
	if err != nil {
		return []sdk.AccAddress{}
	}
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddress)}

	if !bytes.Equal(msg.ValidatorAddress.Bytes(), msg.ValidatorPubKey.Address().Bytes()) {
		// TODO : support to delegate we need to change valAddress.
	}
	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		ValidatorAddress: msg.ValidatorAddress,
		ValidatorPubKey:  sdk.MustBech32ifyValPub(msg.ValidatorPubKey),
		ConsPubKey:       sdk.MustBech32ifyConsPub(msg.ConsPubKey),
		Description:      msg.Description,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface to provide custom
// JSON deserialization of the MsgCreateValidator type.
func (msg *MsgCreateValidator) UnmarshalJSON(bz []byte) error {
	var msgCreateValJSON msgCreateValidatorJSON
	if err := json.Unmarshal(bz, &msgCreateValJSON); err != nil {
		return err
	}

	msg.Description = msgCreateValJSON.Description
	msg.ValidatorAddress = msgCreateValJSON.ValidatorAddress
	var err error
	msg.ValidatorPubKey, err = sdk.GetValPubKeyBech32(msgCreateValJSON.ValidatorPubKey)
	if err != nil {
		return err
	}
	msg.ConsPubKey, err = sdk.GetConsPubKeyBech32(msgCreateValJSON.ConsPubKey)
	if err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !bytes.Equal(msg.ValidatorAddress.Bytes(), msg.ValidatorPubKey.Address().Bytes()) {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}

	return nil
}

//______________________________________________________________________
type MsgBond struct {
	MsgExecute MsgExecute `json:"msg_execute"`
}

// NewMsgBond is a constructor function for MsgSetName
func NewMsgBond(
	tokenContractAddress string,
	fromPubkey secp256k1.PubKeySecp256k1,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte, gasPrice uint64,
	signer sdk.AccAddress,
) MsgBond {
	return MsgBond{
		MsgExecute: MsgExecute{
			ContractAddress: tokenContractAddress,
			ExecPubkey:      fromPubkey,
			SessionCode:     sessionCode,
			SessionArgs:     sessionArgs,
			PaymentCode:     paymentCode,
			PaymentArgs:     paymentArgs,
			GasPrice:        gasPrice,
			Signer:          signer,
		},
	}
}

// Route should return the name of the module
func (msg MsgBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBond) ValidateBasic() sdk.Error {
	if msg.MsgExecute.Signer.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.MsgExecute.Signer}
}

//______________________________________________________________________
type MsgUnBond struct {
	MsgExecute MsgExecute `json:"msg_execute"`
}

// NewMsgUnBond is a constructor function for MsgSetName
func NewMsgUnBond(
	tokenContractAddress string,
	fromPubkey secp256k1.PubKeySecp256k1,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte, gasPrice uint64,
	signer sdk.AccAddress,
) MsgUnBond {
	return MsgUnBond{
		MsgExecute: MsgExecute{
			ContractAddress: tokenContractAddress,
			ExecPubkey:      fromPubkey,
			SessionCode:     sessionCode,
			SessionArgs:     sessionArgs,
			PaymentCode:     paymentCode,
			PaymentArgs:     paymentArgs,
			GasPrice:        gasPrice,
			Signer:          signer,
		},
	}
}

// Route should return the name of the module
func (msg MsgUnBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUnBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUnBond) ValidateBasic() sdk.Error {
	if msg.MsgExecute.Signer.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgUnBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUnBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.MsgExecute.Signer}
}
