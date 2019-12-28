package types

import (
	"bytes"
	"encoding/json"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

// MsgExecute for sending deploy to execution engine
type MsgExecute struct {
	BlockHash            []byte         `json:"block_hash"`
	ExecAccount          sdk.AccAddress `json:"exec_account"`
	ContractOwnerAccount sdk.AccAddress `json:"contract_owner_account"`
	SessionCode          []byte         `json:"session_code"`
	SessionArgs          []byte         `json:"session_args"`
	PaymentCode          []byte         `json:"payment_code"`
	PaymentArgs          []byte         `json:"payment_args"`
	GasPrice             uint64         `json:"gas_price"`
}

// NewMsgExecute is a constructor function for MsgSetName
func NewMsgExecute(
	blockHash []byte,
	execAccount sdk.AccAddress, contractOwnerAccount sdk.AccAddress,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte,
	gasPrice uint64,
) MsgExecute {
	return MsgExecute{
		BlockHash:            blockHash,
		ExecAccount:          execAccount,
		ContractOwnerAccount: contractOwnerAccount,
		SessionCode:          sessionCode,
		SessionArgs:          sessionArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
		GasPrice:             gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgExecute) Route() string { return RouterKey }

// Type should return the action
func (msg MsgExecute) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgExecute) ValidateBasic() sdk.Error {
	if msg.ExecAccount.Equals(sdk.AccAddress("")) || msg.ContractOwnerAccount.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.ExecAccount}
}

// MsgTransfer for sending deploy to execution engine
type MsgTransfer struct {
	TokenOwnerAccount sdk.AccAddress `json:"token_owner_account"`
	FromAccount       sdk.AccAddress `json:"from_account"`
	ToAccount         sdk.AccAddress `json:"to_account"`
	TransferCode      []byte         `json:"transfer_code"`
	TransferArgs      []byte         `json:"transfer_args"`
	PaymentCode       []byte         `json:"payment_code"`
	PaymentArgs       []byte         `json:"payment_args"`
	GasPrice          uint64         `json:"gas_price"`
}

// NewMsgTransfer is a constructor function for MsgSetName
func NewMsgTransfer(
	tokenOwnerAccount sdk.AccAddress,
	fromAccount, toAccount sdk.AccAddress,
	transferCode, transferArgs, paymentCode, paymentArgs []byte,
	gasPrice uint64,
) MsgTransfer {
	return MsgTransfer{
		TokenOwnerAccount: tokenOwnerAccount,
		FromAccount:       fromAccount,
		ToAccount:         toAccount,
		TransferCode:      transferCode,
		TransferArgs:      transferArgs,
		PaymentCode:       paymentCode,
		PaymentArgs:       paymentArgs,
		GasPrice:          gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgTransfer) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTransfer) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if msg.FromAccount.Equals(sdk.AccAddress("")) || msg.TokenOwnerAccount.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.FromAccount}
}

//______________________________________________________________________
// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description      Description    `json:"description" yaml:"description"`
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	PubKey           crypto.PubKey  `json:"pubkey" yaml:"pubkey"`
}

type msgCreateValidatorJSON struct {
	Description      Description    `json:"description" yaml:"description"`
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	PubKey           string         `json:"pubkey" yaml:"pubkey"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddr sdk.ValAddress, pubKey crypto.PubKey, amount uint64,
	description Description,
) MsgCreateValidator {
	return MsgCreateValidator{
		Description:      description,
		DelegatorAddress: sdk.AccAddress(valAddr),
		ValidatorAddress: valAddr,
		PubKey:           pubKey,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.DelegatorAddress}

	if !bytes.Equal(msg.DelegatorAddress.Bytes(), msg.ValidatorAddress.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddress))
	}
	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		Description:      msg.Description,
		DelegatorAddress: msg.DelegatorAddress,
		ValidatorAddress: msg.ValidatorAddress,
		PubKey:           sdk.MustBech32ifyConsPub(msg.PubKey),
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
	msg.DelegatorAddress = msgCreateValJSON.DelegatorAddress
	msg.ValidatorAddress = msgCreateValJSON.ValidatorAddress
	var err error
	msg.PubKey, err = sdk.GetConsPubKeyBech32(msgCreateValJSON.PubKey)
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
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !sdk.AccAddress(msg.ValidatorAddress).Equals(msg.DelegatorAddress) {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}

	return nil
}

//______________________________________________________________________
type MsgBond struct {
	ExecAccount          sdk.AccAddress `json:"exec_account"`
	ContractOwnerAccount sdk.AccAddress `json:"contract_owner_account"`
	SessionCode          []byte         `json:"session_code"`
	SessionArgs          []byte         `json:"session_args"`
	PaymentCode          []byte         `json:"payment_code"`
	PaymentArgs          []byte         `json:"payment_args"`
	GasPrice             uint64         `json:"gas_price"`
}

// NewMsgBond is a constructor function for MsgSetName
func NewMsgBond(
	execAccount sdk.AccAddress, contractOwnerAccount sdk.AccAddress,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte,
	gasPrice uint64,
) MsgBond {
	return MsgBond{
		ExecAccount:          execAccount,
		ContractOwnerAccount: contractOwnerAccount,
		SessionCode:          sessionCode,
		SessionArgs:          sessionArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
		GasPrice:             gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBond) ValidateBasic() sdk.Error {
	if msg.ExecAccount.Equals(sdk.AccAddress("")) || msg.ContractOwnerAccount.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.ExecAccount}
}

//______________________________________________________________________
type MsgUnBond struct {
	ExecAccount          sdk.AccAddress `json:"exec_account"`
	ContractOwnerAccount sdk.AccAddress `json:"contract_owner_account"`
	SessionCode          []byte         `json:"session_code"`
	SessionArgs          []byte         `json:"session_args"`
	PaymentCode          []byte         `json:"payment_code"`
	PaymentArgs          []byte         `json:"payment_args"`
	GasPrice             uint64         `json:"gas_price"`
}

// NewMsgBond is a constructor function for MsgSetName
func NewMsgUnBond(
	execAccount sdk.AccAddress, contractOwnerAccount sdk.AccAddress,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte,
	gasPrice uint64,
) MsgUnBond {
	return MsgUnBond{
		ExecAccount:          execAccount,
		ContractOwnerAccount: contractOwnerAccount,
		SessionCode:          sessionCode,
		SessionArgs:          sessionArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
		GasPrice:             gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgUnBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUnBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUnBond) ValidateBasic() sdk.Error {
	if msg.ExecAccount.Equals(sdk.AccAddress("")) || msg.ContractOwnerAccount.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.ExecAccount}
}
