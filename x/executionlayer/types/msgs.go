package types

import (
	"bytes"
	"encoding/json"

	secp256k1 "github.com/hdac-io/tendermint/crypto/secp256k1"

	sdk "github.com/hdac-io/friday/types"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

// MsgExecute for sending deploy to execution engine
type MsgExecute struct {
	BlockHash       []byte                    `json:"block_hash"`
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
	blockHash []byte,
	contractAddress string,
	execPubkey secp256k1.PubKeySecp256k1,
	sessionCode, sessionArgs []byte,
	paymentCode, paymentArgs []byte,
	gasPrice uint64,
	signer sdk.AccAddress,
) MsgExecute {
	return MsgExecute{
		BlockHash:       blockHash,
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
	TokenContractAddress string                    `json:"token_contract_address"`
	FromPubkey           secp256k1.PubKeySecp256k1 `json:"from_pubkey"`
	ToPubkey             secp256k1.PubKeySecp256k1 `json:"to_pubkey"`
	TransferCode         []byte                    `json:"transfer_code"`
	TransferArgs         []byte                    `json:"transfer_args"`
	PaymentCode          []byte                    `json:"payment_code"`
	PaymentArgs          []byte                    `json:"payment_args"`
	GasPrice             uint64                    `json:"gas_price"`
	Signer               sdk.AccAddress            `json:"signer"`
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
		TokenContractAddress: tokenContractAddress,
		FromPubkey:           fromPubkey,
		ToPubkey:             toPubkey,
		TransferCode:         transferCode,
		TransferArgs:         transferArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
		GasPrice:             gasPrice,
		Signer:               signer,
	}
}

// Route should return the name of the module
func (msg MsgTransfer) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTransfer) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if msg.Signer.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.Signer}
}

//______________________________________________________________________
// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description      Description               `json:"description" yaml:"description"`
	DelegatorAddress sdk.AccAddress            `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress            `json:"validator_address" yaml:"validator_address"`
	PubKey           secp256k1.PubKeySecp256k1 `json:"pubkey" yaml:"pubkey"`
}

type msgCreateValidatorJSON struct {
	Description      Description    `json:"description" yaml:"description"`
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	PubKey           string         `json:"pubkey" yaml:"pubkey"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddr sdk.ValAddress, pubKey secp256k1.PubKeySecp256k1,
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
	ptrPubkey, err := sdk.GetSecp256k1FromRawHexString(msgCreateValJSON.PubKey)
	msg.PubKey = *ptrPubkey
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
	TokenContractAddress string                    `json:"token_contract_address"`
	FromPubkey           secp256k1.PubKeySecp256k1 `json:"from_pubkey"`
	ValAddress           sdk.ValAddress            `json:"val_address"`
	SessionCode          []byte                    `json:"session_code"`
	SessionArgs          []byte                    `json:"session_args"`
	PaymentCode          []byte                    `json:"payment_code"`
	PaymentArgs          []byte                    `json:"payment_args"`
	GasPrice             uint64                    `json:"gas_price"`
	Signer               sdk.AccAddress            `json:"signer"`
}

// NewMsgBond is a constructor function for MsgSetName
func NewMsgBond(
	tokenContractAddress string,
	fromPubkey secp256k1.PubKeySecp256k1, valAddress sdk.ValAddress,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte, gasPrice uint64,
	signer sdk.AccAddress,
) MsgBond {
	return MsgBond{
		TokenContractAddress: tokenContractAddress,
		FromPubkey:           fromPubkey,
		ValAddress:           valAddress,
		SessionCode:          sessionCode,
		SessionArgs:          sessionArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
		GasPrice:             gasPrice,
		Signer:               signer,
	}
}

// Route should return the name of the module
func (msg MsgBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBond) ValidateBasic() sdk.Error {
	if msg.Signer.Equals(sdk.AccAddress("")) || msg.ValAddress.Equals(sdk.ValAddress("")) {
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
	return []sdk.AccAddress{msg.Signer}
}

//______________________________________________________________________
type MsgUnBond struct {
	TokenContractAddress string                    `json:"token_contract_address"`
	FromPubkey           secp256k1.PubKeySecp256k1 `json:"from_pubkey"`
	ValAddress           sdk.ValAddress            `json:"val_address"`
	SessionCode          []byte                    `json:"session_code"`
	SessionArgs          []byte                    `json:"session_args"`
	PaymentCode          []byte                    `json:"payment_code"`
	PaymentArgs          []byte                    `json:"payment_args"`
	GasPrice             uint64                    `json:"gas_price"`
	Signer               sdk.AccAddress            `json:"signer"`
}

// NewMsgUnBond is a constructor function for MsgSetName
func NewMsgUnBond(
	tokenContractAddress string,
	fromPubkey secp256k1.PubKeySecp256k1, valAddress sdk.ValAddress,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte, gasPrice uint64,
	signer sdk.AccAddress,
) MsgBond {
	return MsgBond{
		TokenContractAddress: tokenContractAddress,
		FromPubkey:           fromPubkey,
		ValAddress:           valAddress,
		SessionCode:          sessionCode,
		SessionArgs:          sessionArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
		GasPrice:             gasPrice,
		Signer:               signer,
	}
}

// Route should return the name of the module
func (msg MsgUnBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUnBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUnBond) ValidateBasic() sdk.Error {
	if msg.Signer.Equals(sdk.AccAddress("")) || msg.ValAddress.Equals(sdk.ValAddress("")) {
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
	return []sdk.AccAddress{msg.Signer}
}
