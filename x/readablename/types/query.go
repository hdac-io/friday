package types

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
)

// QueryResUnitAccount payload for a UnitAccount query
type QueryResUnitAccount struct {
	ID      string         `json:"id"`
	Address sdk.AccAddress `json:"address"`
}

// implement fmt.Stringer
func (r QueryResUnitAccount) String() string {
	return fmt.Sprintf("ID: %s\nAddress: %s", r.ID, r.Address)
}
