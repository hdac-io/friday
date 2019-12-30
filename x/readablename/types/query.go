package types

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
)

// QueryReqUnitAccount payload for a UnitAccount query
type QueryReqUnitAccount struct {
	Name string `json:"name"`
}

// QueryResUnitAccount is response of a UnitAccount query
type QueryResUnitAccount struct {
	Name    string         `json:"name"`
	Address sdk.AccAddress `json:"address"`
	PubKey  crypto.PubKey  `json:"pubkey"`
}

// implement fmt.Stringer
func (r QueryResUnitAccount) String() string {
	return fmt.Sprintf("ID: %s\nAddress: %s\nPubkey: %s", r.Name, r.Address, r.PubKey)
}
