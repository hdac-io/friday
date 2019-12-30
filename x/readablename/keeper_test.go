package readablename

import (
	"testing"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/assert"
)

func TestStoreAddDuplicate(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())

	added := store.SetUnitAccount(input.ctx, "bryanrhee", addr, pubkey)
	assert.True(added)

	// cant add twice
	added = store.SetUnitAccount(input.ctx, "bryanrhee", addr, pubkey)
	assert.False(added)
}

func TestStoreKeyChange(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	store.SetUnitAccount(input.ctx, "bryanrhee", addr, pubkey)

	newpubkey := secp256k1.GenPrivKey().PubKey()
	newaddr := sdk.AccAddress(pubkey.Address())
	changed := store.ChangeKey(input.ctx, "bryanrhee", addr, newaddr, pubkey, newpubkey)
	assert.True(changed)
}

func TestStoreAddrCheck(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	store.SetUnitAccount(input.ctx, "bryanrhee", addr, pubkey)

	verified := store.AddrCheck(input.ctx, "bryanrhee", addr)
	assert.True(verified)

	notverified := store.AddrCheck(input.ctx, "bryanrh", addr)
	assert.False(notverified)
}
