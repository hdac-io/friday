package types

import (
	sdk "github.com/hdac-io/friday/types"
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
