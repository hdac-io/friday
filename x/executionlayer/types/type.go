package types

import (
	"fmt"
	"strings"
)

// UnitHashMap used to define Unit account structure
type UnitHashMap struct {
	BlockState []byte `json:"block_state"`
	EEState    []byte `json:"ee_state"`
}

// NewUnitHashMap returns a new UnitAccount
func NewUnitHashMap() UnitHashMap {
	return UnitHashMap{}
}

// implement fmt.Stringer
func (u UnitHashMap) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Block state: %s
EE state: %s`, u.BlockState, u.EEState))
}
