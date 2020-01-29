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
type MsgSetNickname struct {
	Nickname Name           `json:"nickname"`
	Address  sdk.AccAddress `json:"address"`
}

// NewMsgSetNickname is a constructor function for MsgSetName
func NewMsgSetNickname(name Name, address sdk.AccAddress) MsgSetNickname {
	return MsgSetNickname{
		Nickname: name,
		Address:  address,
	}
}

// Route should return the name of the module
func (msg MsgSetNickname) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetNickname) Type() string { return "newaccount" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetNickname) ValidateBasic() sdk.Error {
	if msg.Address.String() == "" {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if msg.Nickname.Equal(NewName("")) {
		return sdk.ErrUnknownRequest("ID cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetNickname) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetNickname) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}

///////////////////////////////////
////////// Change Key /////////////
///////////////////////////////////

// MsgChangeKey defines a ChangeKey message
type MsgChangeKey struct {
	Nickname   string         `json:"nickname"`
	OldAddress sdk.AccAddress `json:"old_address"`
	NewAddress sdk.AccAddress `json:"new_address"`
}

// NewMsgChangeKey is a constructor function for MsgChangeKey
func NewMsgChangeKey(name string,
	oldAddress, newAddress sdk.AccAddress,
) MsgChangeKey {
	return MsgChangeKey{
		Nickname:   name,
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
	if len(msg.OldAddress.Bytes()) == 0 || len(msg.NewAddress.Bytes()) == 0 {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if len(msg.Nickname) == 0 {
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
	return []sdk.AccAddress{msg.OldAddress}
}
