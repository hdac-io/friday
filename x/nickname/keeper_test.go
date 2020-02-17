package nickname

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

	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	added := store.SetNickname(input.ctx, "bryanrhee", addr)
	assert.True(added)

	// cant add twice
	added = store.SetNickname(input.ctx, "bryanrhee", addr)
	assert.False(added)
}

func TestStoreKeyChange(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	store.SetNickname(input.ctx, "bryanrhee", addr)

	newaddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	changed := store.ChangeKey(input.ctx, "bryanrhee", addr, newaddr)
	assert.True(changed)
}

func TestStoreAddrCheck(t *testing.T) {
	assert := assert.New(t)
	input := setupTestInput()
	store := input.k

	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	store.SetNickname(input.ctx, "bryanrhee", addr)

	verified := store.AddrCheck(input.ctx, "bryanrhee", addr)
	assert.True(verified)

	notverified := store.AddrCheck(input.ctx, "bryanrh", addr)
	assert.False(notverified)
}
