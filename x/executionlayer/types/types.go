package types

import (
	"fmt"
	"strings"
)

var (
	SYSTEM_ACCOUNT   = make([]byte, 32)
	TEMP_ACC_ADDRESS = make([]byte, 20)
)

const (
	MintContractName = "mint"
	PosContractName  = "pos"

	ProxyContractName         = "client_api_proxy"
	TransferMethodName        = "transfer_to_account"
	PaymentMethodName         = "standard_payment"
	BondMethodName            = "bond"
	UnbondMethodName          = "unbond"
	DelegateMethodName        = "delegate"
	UndelegateMethodName      = "undelegate"
	RedelegateMethodName      = "redelegate"
	VoteMethodName            = "vote"
	UnvoteMethodName          = "unvote"
	StepMethodName            = "step"
	ClaimRewardMethodName     = "claim_reward"
	ClaimCommissionMethodName = "claim_commission"

	SYSTEM_ACCOUNT_BALANCE       = "1000000000000000000000000000000"
	TRANSFER_BALANCE             = "999999999999000000000000000000"
	SYSTEM_ACCOUNT_BONDED_AMOUNT = "0"
	BASIC_FEE                    = "10000000000000000"
	BASIC_GAS                    = 30000000

	DECIMAL_POINT_POS = 18

	RewardString     = "reward"
	CommissionString = "commission"
	RewardValue      = true
	CommissionValue  = false
)

// UnitHashMap used to define Unit account structure
type UnitHashMap struct {
	EEState []byte `json:"ee_state"`
}

// NewUnitHashMap returns a new UnitAccount
func NewUnitHashMap(eeState []byte) UnitHashMap {
	return UnitHashMap{
		EEState: eeState,
	}
}

// implement fmt.Stringer
func (u UnitHashMap) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Unit Hash
  EE state: %s`, u.EEState))
}
