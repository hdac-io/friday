package types

import (
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// RouterKey is not in sense yet
const RouterKey = ModuleName

////////////////////////////
/////// Add Account ////////
////////////////////////////

// MsgSetAccount defines a SetAccount message
type MsgSetAccount struct {
	Name    Name           `json:"name"`
	Address sdk.AccAddress `json:"address"`
	PubKey  crypto.PubKey  `json:"pubkey"`
}

// NewMsgSetAccount is a constructor function for MsgSetName
func NewMsgSetAccount(name Name, address sdk.AccAddress, pubkey crypto.PubKey) MsgSetAccount {
	return MsgSetAccount{
		Name:    name,
		Address: address,
		PubKey:  pubkey,
	}
}

// Route should return the name of the module
func (msg MsgSetAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetAccount) Type() string { return "newaccount" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetAccount) ValidateBasic() sdk.Error {
	if msg.PubKey.Address().String() == "" {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if msg.Name.Equal(NewName("")) {
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
////////// Change Key /////////////
///////////////////////////////////

// MsgChangeKey defines a ChangeKey message
type MsgChangeKey struct {
	ID         string         `json:"ID"`
	OldAddress sdk.AccAddress `json:"old_address"`
	NewAddress sdk.AccAddress `json:"new_address"`
	OldPubKey  crypto.PubKey  `json:"old_pubkey"`
	NewPubKey  crypto.PubKey  `json:"new_pubkey"`
}

// NewMsgChangeKey is a constructor function for MsgChangeKey
func NewMsgChangeKey(name string,
	oldAddress, newAddress sdk.AccAddress,
	oldPubKey, newPubKey crypto.PubKey) MsgChangeKey {
	return MsgChangeKey{
		ID:         name,
		OldAddress: oldAddress,
		NewAddress: newAddress,
		OldPubKey:  oldPubKey,
		NewPubKey:  newPubKey,
	}
}

// Route should return the name of the module
func (msg MsgChangeKey) Route() string { return RouterKey }

// Type should return the action
func (msg MsgChangeKey) Type() string { return "changekey" }

// ValidateBasic runs stateless checks on the message
func (msg MsgChangeKey) ValidateBasic() sdk.Error {
	if len(msg.OldPubKey.Bytes()) == 0 || len(msg.NewPubKey.Bytes()) == 0 {
		return sdk.ErrUnknownRequest("PubKey cannot be empty")
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
	return []sdk.AccAddress{msg.OldAddress}
}
