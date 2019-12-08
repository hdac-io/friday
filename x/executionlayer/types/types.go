package types

import (
	"encoding/base64"
	"fmt"
	"strings"

	sdk "github.com/hdac-io/friday/types"
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

// PublicKey for Execution Engines
type PublicKey []byte

// NewPublicKey is constructor for PublicKey,
// accepts base64 encoded public key string and returns PublicKey
func NewPublicKey(base64PublicKey string) (*PublicKey, error) {
	publicKey, err := base64.StdEncoding.DecodeString(base64PublicKey)
	if err != nil || len(publicKey) != 32 {
		return nil, ErrPublicKeyDecode(DefaultCodespace, base64PublicKey)
	}
	ret := PublicKey(publicKey)
	return &ret, nil
}

// ToPublicKey convert sdk.AccAddress to PublicKey appending null padding.
// we currently use sdk.AccAddress as public key for PoC.
// This should be removed later.
func ToPublicKey(accAddr sdk.AccAddress) PublicKey {
	return append(accAddr.Bytes(), make([]byte, 12)...)
}
