package types

import (
	"fmt"
	"strings"
)

var (
	SYSTEM_ACCOUNT = make([]byte, 32)
)

const (
	MintContractName = "mint"
	PosContractName  = "pos"

	ProxyContractName    = "client_api_proxy"
	TransferMethodName   = "transfer_to_account"
	PaymentMethodName    = "standard_payment"
	BondMethodName       = "bond"
	UnbondMethodName     = "unbond"
	DelegateMethodName   = "delegate"
	UndelegateMethodName = "undelegate"
	RedelegateMethodName = "redelegate"
)

// UnitHashMap used to define Unit account structure
type UnitHashMap struct {
	EEState []byte `json:"ee_state"`
}

// NewUnitHashMap returns a new UnitAccount
func NewUnitHashMap() UnitHashMap {
	return UnitHashMap{}
}

// implement fmt.Stringer
func (u UnitHashMap) String() string {
	return strings.TrimSpace(fmt.Sprintf(`EE state: %s`, u.EEState))
}
