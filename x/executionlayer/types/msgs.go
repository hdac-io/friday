package types

import (
	"encoding/json"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

// MsgExecute for sending deploy to execution engine
type MsgExecute struct {
	ContractAddress string            `json:"contract_address"`
	ExecAddress     sdk.AccAddress    `json:"exec_address"`
	SessionType     util.ContractType `json:"session_type"`
	SessionCode     []byte            `json:"session_code"`
	SessionArgs     string            `json:"session_args"`
	Fee             string            `json:"fee"`
}

// NewMsgExecute is a constructor function for MsgSetName
func NewMsgExecute(
	contractAddress string,
	execAddress sdk.AccAddress,
	sessionType util.ContractType,
	sessionCode []byte,
	sessionArgs string,
	fee string,
) MsgExecute {
	return MsgExecute{
		ExecAddress:     execAddress,
		ContractAddress: contractAddress,
		SessionType:     sessionType,
		SessionCode:     sessionCode,
		SessionArgs:     sessionArgs,
		Fee:             fee,
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
	Amount          string         `json:"amount" yaml:"amount"`
	Fee             string         `json:"fee" yaml:"fee"`
}

// NewMsgTransfer is a constructor function for MsgSetName
func NewMsgTransfer(
	tokenContractAddress string,
	fromAddress, toAddress sdk.AccAddress,
	amount, fee string,
) MsgTransfer {
	return MsgTransfer{
		ContractAddress: tokenContractAddress,
		FromAddress:     fromAddress,
		ToAddress:       toAddress,
		Amount:          amount,
		Fee:             fee,
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
	Amount          string         `json:"amount" yaml:"amount"`
	Fee             string         `json:"fee" yaml:"fee"`
}

// NewMsgBond is a constructor function for MsgSetName
func NewMsgBond(
	tokenContractAddress string,
	bonderAddress sdk.AccAddress,
	amount, fee string,
) MsgBond {
	return MsgBond{
		ContractAddress: tokenContractAddress,
		FromAddress:     bonderAddress,
		Amount:          amount,
		Fee:             fee,
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
	Amount          string         `json:"amount" yaml:"amount"`
	Fee             string         `json:"fee" yaml:"fee"`
}

// NewMsgUnBond is a constructor function for MsgSetName
func NewMsgUnBond(
	tokenContractAddress string,
	unbonderAddress sdk.AccAddress,
	amount, fee string,
) MsgUnBond {
	return MsgUnBond{
		ContractAddress: tokenContractAddress,
		FromAddress:     unbonderAddress,
		Amount:          amount,
		Fee:             fee,
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

//______________________________________________________________________
type MsgDelegate struct {
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	FromAddress     sdk.AccAddress `json:"from_address" yaml:"from_address"`
	ValAddress      sdk.AccAddress `json:"val_address" yaml:"val_address"`
	Amount          string         `json:"amount" yaml:"amount"`
	Fee             string         `json:"fee" yaml:"fee"`
}

// NewMsgDelegate is a constructor function for MsgSetName
func NewMsgDelegate(
	tokenContractAddress string,
	fromAddress, vaildatorAddress sdk.AccAddress,
	amount, fee string,
) MsgDelegate {
	return MsgDelegate{
		ContractAddress: tokenContractAddress,
		FromAddress:     fromAddress,
		ValAddress:      vaildatorAddress,
		Amount:          amount,
		Fee:             fee,
	}
}

// Route should return the name of the module
func (msg MsgDelegate) Route() string { return RouterKey }

// Type should return the action
func (msg MsgDelegate) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
type MsgUndelegate struct {
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	FromAddress     sdk.AccAddress `json:"from_address" yaml:"from_address"`
	ValAddress      sdk.AccAddress `json:"val_address" yaml:"val_address"`
	Amount          string         `json:"amount" yaml:"amount"`
	Fee             string         `json:"fee" yaml:"fee"`
}

// NewMsgUndelegate is a constructor function for MsgSetName
func NewMsgUndelegate(
	tokenContractAddress string,
	fromAddress, vaildatorAddress sdk.AccAddress,
	amount, fee string,
) MsgUndelegate {
	return MsgUndelegate{
		ContractAddress: tokenContractAddress,
		FromAddress:     fromAddress,
		ValAddress:      vaildatorAddress,
		Amount:          amount,
		Fee:             fee,
	}
}

// Route should return the name of the module
func (msg MsgUndelegate) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUndelegate) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUndelegate) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgUndelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
type MsgRedelegate struct {
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	FromAddress     sdk.AccAddress `json:"from_address" yaml:"from_address"`
	SrcValAddress   sdk.AccAddress `json:"src_val_address" yaml:"src_val_address"`
	DestValAddress  sdk.AccAddress `json:"dest_val_address" yaml:"dest_val_address"`
	Amount          string         `json:"amount" yaml:"amount"`
	Fee             string         `json:"fee" yaml:"fee"`
}

// MsgRedelegate is a constructor function for MsgSetName
func NewMsgRedelegate(
	tokenContractAddress string,
	fromAddress, srcVaildatorAddress, descVaildatorAddress sdk.AccAddress,
	amount, fee string,
) MsgRedelegate {
	return MsgRedelegate{
		ContractAddress: tokenContractAddress,
		FromAddress:     fromAddress,
		SrcValAddress:   srcVaildatorAddress,
		DestValAddress:  descVaildatorAddress,
		Amount:          amount,
		Fee:             fee,
	}
}

// Route should return the name of the module
func (msg MsgRedelegate) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRedelegate) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgRedelegate) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgRedelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRedelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
type MsgVote struct {
	ContractAddress       string              `json:"contract_address" yaml:"contract_address"`
	FromAddress           sdk.AccAddress      `json:"from_address" yaml:"from_address"`
	TargetContractAddress sdk.ContractAddress `json:"target_contract_address" yaml:"target_contract_address"`
	Amount                string              `json:"amount" yaml:"amount"`
	Fee                   string              `json:"fee" yaml:"fee"`
}

// NewMsgVote is a constructor function for MsgSetName
func NewMsgVote(
	tokenContractAddress string,
	fromAddress sdk.AccAddress,
	targetContractAddress sdk.ContractAddress,
	amount, fee string,
) MsgVote {
	return MsgVote{
		ContractAddress:       tokenContractAddress,
		FromAddress:           fromAddress,
		TargetContractAddress: targetContractAddress,
		Amount:                amount,
		Fee:                   fee,
	}
}

// Route should return the name of the module
func (msg MsgVote) Route() string { return RouterKey }

// Type should return the action
func (msg MsgVote) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgVote) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if len(msg.TargetContractAddress.Bytes()) != 32 {
		return sdk.ErrUnknownRequest("Hash must be 32 bytes")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
type MsgUnvote struct {
	ContractAddress       string              `json:"contract_address" yaml:"contract_address"`
	FromAddress           sdk.AccAddress      `json:"from_address" yaml:"from_address"`
	TargetContractAddress sdk.ContractAddress `json:"target_contract_address" yaml:"target_contract_address"`
	Amount                string              `json:"amount" yaml:"amount"`
	Fee                   string              `json:"fee" yaml:"fee"`
}

// NewMsgUnvote is a constructor function for MsgSetName
func NewMsgUnvote(
	tokenContractAddress string,
	fromAddress sdk.AccAddress,
	targetContractAddress sdk.ContractAddress,
	amount, fee string,
) MsgUnvote {
	return MsgUnvote{
		ContractAddress:       tokenContractAddress,
		FromAddress:           fromAddress,
		TargetContractAddress: targetContractAddress,
		Amount:                amount,
		Fee:                   fee,
	}
}

// Route should return the name of the module
func (msg MsgUnvote) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUnvote) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUnvote) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if len(msg.TargetContractAddress.Bytes()) != 32 {
		return sdk.ErrUnknownRequest("Hash must be 32 bytes")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgUnvote) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUnvote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

//______________________________________________________________________
type MsgClaim struct {
	ContractAddress    string         `json:"contract_address" yaml:"contract_address"`
	FromAddress        sdk.AccAddress `json:"from_address" yaml:"from_address"`
	RewardOrCommission bool           `json:"reward_or_commission" yaml:"reward_or_commission"`
	Fee                string         `json:"fee" yaml:"fee"`
}

// NewMsgClaim is a constructor function for MsgSetName
func NewMsgClaim(
	tokenContractAddress string,
	fromAddress sdk.AccAddress,
	rewardOrCommission bool,
	fee string,
) MsgClaim {
	return MsgClaim{
		ContractAddress:    tokenContractAddress,
		FromAddress:        fromAddress,
		RewardOrCommission: rewardOrCommission,
		Fee:                fee,
	}
}

// Route should return the name of the module
func (msg MsgClaim) Route() string { return RouterKey }

// Type should return the action
func (msg MsgClaim) Type() string { return "executionengine" }

// ValidateBasic runs stateless checks on the message
func (msg MsgClaim) ValidateBasic() sdk.Error {
	if msg.FromAddress.Equals(sdk.AccAddress("")) {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}
