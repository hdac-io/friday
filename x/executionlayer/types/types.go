package types

import (
	"fmt"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	sdk "github.com/hdac-io/friday/types"
)

// UnitHashMap used to define Unit account structure
type UnitHashMap struct {
	EEState []byte `json:"ee_state"`
}

type CandidateBlock struct {
	Hash  []byte      `json:"hash"`
	Bonds []*ipc.Bond `json:"bonds"`
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

// ToPublicKey convert sdk.AccAddress to PublicKey appending null padding.
// we currently use sdk.AccAddress as public key for PoC.
// This should be removed later.
// TODO: Replace fridayvaloper as Secp256k1-like conversion in handler.go/Endblocker() and delete the type
func ToPublicKey(addr sdk.Address) PublicKey {
	return append(addr.Bytes(), make([]byte, 12)...)
}
