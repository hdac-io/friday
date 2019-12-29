package types

import (
	"fmt"
	"strings"

	sdk "github.com/hdac-io/friday/types"
)

// UnitAccount used to define Unit account structure
type UnitAccount struct {
	ID      Name           `json:"id"`
	Address sdk.AccAddress `json:"address"`
	// To be more appendded..
}

// NewUnitAccount returns a new UnitAccount
func NewUnitAccount() UnitAccount {
	return UnitAccount{}
}

// implement fmt.Stringer
func (w UnitAccount) String() string {
	strname, _ := w.ID.ToString()
	return strings.TrimSpace(fmt.Sprintf(`ID: %s
Address: %s`, strname, w.Address))
}
