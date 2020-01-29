package types

import (
	"fmt"
	"strings"

	sdk "github.com/hdac-io/friday/types"
)

// UnitAccount used to define Unit account structure
type UnitAccount struct {
	Nickname Name           `json:"nick"`
	Address  sdk.AccAddress `json:"address"`
}

// NewUnitAccount returns a new UnitAccount
func NewUnitAccount(name Name, address sdk.AccAddress) UnitAccount {
	return UnitAccount{
		Nickname: name,
		Address:  address,
	}
}

// implement fmt.Stringer
func (w UnitAccount) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Nick: %s
Address: %s`, w.Nickname.MustToString(), w.Address))
}
