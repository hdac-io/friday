package types

import (
	"encoding/json"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

// MsgExecute for sending deploy to execution engine
type MsgExecute struct {
	ContractAddress string                  `json:"contract_address"`
	ExecAddress     sdk.AccAddress          `json:"exec_address"`
	SessionType     util.ContractType       `json:"session_type"`
	SessionCode     []byte                  `json:"session_code"`
	SessionArgs     []*consensus.Deploy_Arg `json:"session_args"`
	Fee             uint64                  `json:"fee"`
	GasPrice        uint64                  `json:"gas_price"`
}

// NewMsgExecute is a constructor function for MsgSetName
func NewMsgExecute(
	contractAddress string,
	execAddress sdk.AccAddress,
	sessionType util.ContractType,
	sessionCode []byte,
	sessionArgs []*consensus.Deploy_Arg,
	fee, gasPrice uint64,
) MsgExecute {
	return MsgExecute{
		ExecAddress:     execAddress,
		ContractAddress: contractAddress,
		SessionType:     sessionType,
		SessionCode:     sessionCode,
		SessionArgs:     sessionArgs,
		Fee:             fee,
		GasPrice:        gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgExecute) Route() string { return RouterKey }

// Type should return the action
func (msg MsgExecute) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgExecute) ValidateBasic() sdk.Error {
	if msg.ExecAddress.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.ExecAddress}
}

// MsgTransfer for sending deploy to execution engine
type MsgTransfer struct {
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	FromAddress     sdk.AccAddress `json:"from_address" yaml:"from_address"`
	ToAddress       sdk.AccAddress `json:"to_address" yaml:"to_address"`
	Amount          uint64         `json:"amount" yaml:"amount"`
	Fee             uint64         `json:"fee" yaml:"fee"`
	GasPrice        uint64         `json:"gas_price" yaml:"gas_price"`
}

// NewMsgTransfer is a constructor function for MsgSetName
func NewMsgTransfer(
	tokenContractAddress string,
	fromAddress, toAddress sdk.AccAddress,
	amount, fee, gasPrice uint64,
) MsgTransfer {
	return MsgTransfer{
		ContractAddress: tokenContractAddress,
		FromAddress:     fromAddress,
		ToAddress:       toAddress,
		Amount:          amount,
		Fee:             fee,
		GasPrice:        gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgTransfer) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTransfer) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if msg.ToAddress.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	ValidatorAddress sdk.AccAddress `json:"validator_address" yaml:"validator_address"`
	ConsPubKey       crypto.PubKey  `json:"cons_pubkey" yaml:"cons_pubkey"`
	Description      Description    `json:"description" yaml:"description"`
}

type msgCreateValidatorJSON struct {
	ValidatorAddress sdk.AccAddress `json:"validator_address" yaml:"validator_address"`
	ConsPubKey       string         `json:"cons_pubkey" yaml:"cons_pubkey"`
	Description      Description    `json:"description" yaml:"description"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddress sdk.AccAddress,
	consPubKey crypto.PubKey,
	description Description,
) MsgCreateValidator {
	return MsgCreateValidator{
		ValidatorAddress: valAddress,
		ConsPubKey:       consPubKey,
		Description:      description,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.ValidatorAddress}

	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		ValidatorAddress: msg.ValidatorAddress,
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
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}

	return nil
}

//______________________________________________________________________
// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	ValidatorAddress sdk.AccAddress `json:"address" yaml:"address"`
	Description
}

func NewMsgEditValidator(valAddr sdk.AccAddress, description Description) MsgEditValidator {
	return MsgEditValidator{
		ValidatorAddress: valAddr,
		Description:      description,
	}
}

//nolint
func (msg MsgEditValidator) Route() string { return RouterKey }
func (msg MsgEditValidator) Type() string  { return "edit_validator" }
func (msg MsgEditValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.ValidatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddress.Empty() {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}
	return nil
}

//______________________________________________________________________
type MsgBond struct {
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	FromAddress     sdk.AccAddress `json:"from_address" yaml:"from_address"`
	Amount          uint64         `json:"amount" yaml:"amount"`
	Fee             uint64         `json:"fee" yaml:"fee"`
	GasPrice        uint64         `json:"gas_price" yaml:"gas_price"`
}

// NewMsgBond is a constructor function for MsgSetName
func NewMsgBond(
	tokenContractAddress string,
	bonderAddress sdk.AccAddress,
	amount, fee, gasPrice uint64,
) MsgBond {
	return MsgBond{
		ContractAddress: tokenContractAddress,
		FromAddress:     bonderAddress,
		Amount:          amount,
		Fee:             fee,
		GasPrice:        gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBond) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
type MsgUnBond struct {
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	FromAddress     sdk.AccAddress `json:"from_address" yaml:"from_address"`
	Amount          uint64         `json:"amount" yaml:"amount"`
	Fee             uint64         `json:"fee" yaml:"fee"`
	GasPrice        uint64         `json:"gas_price" yaml:"gas_price"`
}

// NewMsgUnBond is a constructor function for MsgSetName
func NewMsgUnBond(
	tokenContractAddress string,
	unbonderAddress sdk.AccAddress,
	amount, fee, gasPrice uint64,
) MsgUnBond {
	return MsgUnBond{
		ContractAddress: tokenContractAddress,
		FromAddress:     unbonderAddress,
		Amount:          amount,
		Fee:             fee,
		GasPrice:        gasPrice,
	}
}

// Route should return the name of the module
func (msg MsgUnBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUnBond) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUnBond) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
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
	return []sdk.AccAddress{msg.FromAddress}
}
