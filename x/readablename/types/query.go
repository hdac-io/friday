package types

import (
	"fmt"

	"github.com/hdac-io/tendermint/crypto/secp256k1"
	sdk "github.com/hdac-io/friday/types"
)

// QueryReqUnitAccount payload for a UnitAccount query
type QueryReqUnitAccount struct {
	Name Name `json:"name"`
}

// QueryResUnitAccount is response of a UnitAccount query
type QueryResUnitAccount struct {
	Name    Name                      `json:"name"`
	Address sdk.AccAddress            `json:"address"`
	PubKey  secp256k1.PubKeySecp256k1 `json:"pubkey"`
}

// implement fmt.Stringer
func (r QueryResUnitAccount) String() string {
	return fmt.Sprintf("ID: %s\nAddress: %s\nPubkey: %s",
		r.Name.MustToString(), r.Address.String(), r.PubKey.String())
}
