package types

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto/secp256k1"
)

// QueryReqUnitAccount payload for a UnitAccount query
type QueryReqUnitAccount struct {
	Name string `json:"name"`
}

// QueryResUnitAccount is response of a UnitAccount query
type QueryResUnitAccount struct {
	Name    string                    `json:"name"`
	Address sdk.AccAddress            `json:"address"`
	PubKey  secp256k1.PubKeySecp256k1 `json:"pubkey"`
}

// implement fmt.Stringer
func (r QueryResUnitAccount) String() string {
	return fmt.Sprintf("ID: %s\nAddress: %s\nPubkey: %s", r.Name, r.Address.String(), r.PubKey.String())
}
