package nameservice

import (
	"math/rand"
	"testing"

	"github.com/hdac-io/friday/types"
	sdk "github.com/hdac-io/friday/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/assert"
)

//-------------------------------------------

func addressGenerator(seed []byte) string {
	var pub ed25519.PubKeyEd25519
	rand.Read(seed)
	acc := types.AccAddress(pub.Address())
	str := acc.String()
	return str
}

func TestStoreAddDuplicate(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	strAddress := addressGenerator([]byte("1"))

	addr, err := sdk.AccAddressFromBech32(strAddress)
	if err != nil {
		panic(err)
	}
	added := store.SetUnitAccount(input.ctx, "bryanrhee", addr)
	assert.True(added)

	// cant add twice
	added = store.SetUnitAccount(input.ctx, "bryanrhee", addr)
	assert.False(added)
}

func TestStoreKeyChange(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	strAddressCurr := addressGenerator([]byte("1"))

	addr, err := sdk.AccAddressFromBech32(strAddressCurr)
	if err != nil {
		panic(err)
	}
	store.SetUnitAccount(input.ctx, "bryanrhee", addr)

	strAddressNew := addressGenerator([]byte("2"))
	// Try to change
	newaddr, err := sdk.AccAddressFromBech32(strAddressNew)
	if err != nil {
		panic(err)
	}
	changed := store.ChangeKey(input.ctx, "bryanrhee", addr, newaddr)
	assert.True(changed)
}

func TestStoreAddrCheck(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k
	strAddress := addressGenerator([]byte("1"))

	addr, err := sdk.AccAddressFromBech32(strAddress)
	if err != nil {
		panic(err)
	}
	store.SetUnitAccount(input.ctx, "bryanrhee", addr)

	verified := store.AddrCheck(input.ctx, "bryanrhee", addr)
	assert.True(verified)

	notverified := store.AddrCheck(input.ctx, "bryanrh", addr)
	assert.False(notverified)
}
