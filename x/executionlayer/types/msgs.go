package types

import (
	sdk "github.com/hdac-io/friday/types"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

// MsgExecute for sending deploy to execution engine
type MsgExecute struct {
	BlockState           []byte         `json:"block_state"`
	ExecAccount          sdk.AccAddress `json:"exec_account"`
	ContractOwnerAccount sdk.AccAddress `json:"contract_owner_account"`
	SessionCode          []byte         `json:"session_code"`
	SessionArgs          []byte         `json:"session_args"`
	PaymentCode          []byte         `json:"payment_code"`
	PaymentArgs          []byte         `json:"payment_args"`
}

// NewMsgExecute is a constructor function for MsgSetName
func NewMsgExecute(
	blockState []byte,
	execAccount sdk.AccAddress, contractOwnerAccount sdk.AccAddress,
	sessionCode []byte, sessionArgs []byte,
	paymentCode []byte, paymentArgs []byte,
) MsgExecute {
	return MsgExecute{
		BlockState:           blockState,
		ExecAccount:          execAccount,
		ContractOwnerAccount: contractOwnerAccount,
		SessionCode:          sessionCode,
		SessionArgs:          sessionArgs,
		PaymentCode:          paymentCode,
		PaymentArgs:          paymentArgs,
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
