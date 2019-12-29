package types

import (
	sdk "github.com/hdac-io/friday/types"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

////////////////////////////
/////// Add Account ////////
////////////////////////////

// MsgSetAccount defines a SetAccount message
type MsgSetAccount struct {
	ID      string         `json:"id"`
	Address sdk.AccAddress `json:"address"`
}

// NewMsgSetAccount is a constructor function for MsgSetName
func NewMsgSetAccount(name string, address sdk.AccAddress) MsgSetAccount {
	return MsgSetAccount{
		ID:      name,
		Address: address,
	}
}

// Route should return the name of the module
func (msg MsgSetAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetAccount) Type() string { return "newaccount" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetAccount) ValidateBasic() sdk.Error {
	if msg.Address.String() == "" {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if len(msg.ID) == 0 {
		return sdk.ErrUnknownRequest("ID cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetAccount) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

///////////////////////////////////
////////// Key check //////////////
///////////////////////////////////

// MsgAddrCheck defines a ChangeKey message
type MsgAddrCheck struct {
	ID      string         `json:"id"`
	Address sdk.AccAddress `json:"address"`
}

// NewMsgAddrCheck is a constructor function for MsgAddrCheck
func NewMsgAddrCheck(name string, address sdk.AccAddress) MsgAddrCheck {
	return MsgAddrCheck{
		ID:      name,
		Address: address,
	}
}

// Route should return the name of the module
func (msg MsgAddrCheck) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAddrCheck) Type() string { return "addrcheck" }

// ValidateBasic runs stateless checks on the message
func (msg MsgAddrCheck) ValidateBasic() sdk.Error {
	if msg.ID == "" || msg.Address.String() == "" {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if len(msg.ID) == 0 {
		return sdk.ErrUnknownRequest("ID cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgAddrCheck) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgAddrCheck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

///////////////////////////////////
////////// Change Key /////////////
///////////////////////////////////

// MsgChangeKey defines a ChangeKey message
type MsgChangeKey struct {
	ID         string         `json:"ID"`
	OldAddress sdk.AccAddress `json:"oldaddress"`
	NewAddress sdk.AccAddress `json:"newaddress"`
}

// NewMsgChangeKey is a constructor function for MsgChangeKey
func NewMsgChangeKey(name string, oldAddress, newAddress sdk.AccAddress) MsgChangeKey {
	return MsgChangeKey{
		ID:         name,
		OldAddress: oldAddress,
		NewAddress: newAddress,
	}
}

// Route should return the name of the module
func (msg MsgChangeKey) Route() string { return RouterKey }

// Type should return the action
func (msg MsgChangeKey) Type() string { return "changekey" }

// ValidateBasic runs stateless checks on the message
func (msg MsgChangeKey) ValidateBasic() sdk.Error {
	if msg.OldAddress.String() == "" || msg.NewAddress.String() == "" {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if len(msg.ID) == 0 {
		return sdk.ErrUnknownRequest("ID cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgChangeKey) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgChangeKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.NewAddress}
}
