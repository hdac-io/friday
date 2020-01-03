package types

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto"
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

// NewPublicKeyFromCryptoPubkey is constructor for PublicKey,
// accepts base64 encoded public key string and returns PublicKey
func NewPublicKeyFromCryptoPubkey(cryptoPubKey crypto.PubKey) *PublicKey {
	ret := PublicKey(cryptoPubKey.Bytes())
	return &ret
}

// ToPublicKey convert sdk.AccAddress to PublicKey appending null padding.
// we currently use sdk.AccAddress as public key for PoC.
// This should be removed later.
func ToPublicKey(addr sdk.Address) PublicKey {
	return append(addr.Bytes(), make([]byte, 12)...)
}
