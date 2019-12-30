package types

import (
	"fmt"
	"strings"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// UnitAccount used to define Unit account structure
type UnitAccount struct {
	Name    Name           `json:"id"`
	Address sdk.AccAddress `json:"address"`
	PubKey  crypto.PubKey  `json:"pubkey"`
	// To be more appendded..
}

// NewUnitAccount returns a new UnitAccount
func NewUnitAccount(name Name, address sdk.AccAddress, pubkey crypto.PubKey) UnitAccount {
	return UnitAccount{
		Name:    name,
		Address: address,
		PubKey:  pubkey,
	}
}

// implement fmt.Stringer
func (w UnitAccount) String() string {
	return strings.TrimSpace(fmt.Sprintf(`ID: %s
Public key: %s
Address: %s`, w.Name.MustToString(), w.PubKey.Address().String(), w.Address))
}
