package types

import (
	"fmt"
	"strings"

	"github.com/hdac-io/tendermint/crypto/secp256k1"

	sdk "github.com/hdac-io/friday/types"
)

// UnitAccount used to define Unit account structure
type UnitAccount struct {
	Name    Name                      `json:"id"`
	Address sdk.AccAddress            `json:"address"`
	PubKey  secp256k1.PubKeySecp256k1 `json:"pubkey"`
	// To be more appendded..
}

// NewUnitAccount returns a new UnitAccount
func NewUnitAccount(name Name, address sdk.AccAddress, pubkey secp256k1.PubKeySecp256k1) UnitAccount {
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
Address: %s`, w.Name.MustToString(), w.PubKey.String(), w.Address))
}
