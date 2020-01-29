package types

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
)

// QueryReqUnitAccount payload for a UnitAccount query
type QueryReqUnitAccount struct {
	Nickname string `json:"nickname"`
}

// QueryResUnitAccount is response of a UnitAccount query
type QueryResUnitAccount struct {
	Nickname string         `json:"nickname"`
	Address  sdk.AccAddress `json:"address"`
}

// implement fmt.Stringer
func (r QueryResUnitAccount) String() string {
	return fmt.Sprintf("Nickname: %s\nAddress: %s", r.Nickname, r.Address.String())
}
