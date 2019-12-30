package readablename

import (
	"testing"

	sdk "github.com/hdac-io/friday/types"

	"github.com/hdac-io/tendermint/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	res := h(input.ctx, NewMsgSetAccount(NewName("bryanrhee"), addr, pubkey))
	require.True(t, res.IsOK())

	newpubkey := secp256k1.GenPrivKey().PubKey()
	newaddr := sdk.AccAddress(newpubkey.Address())
	res = h(input.ctx, NewMsgChangeKey("bryanrhee", addr, newaddr, pubkey, newpubkey))
	require.True(t, res.IsOK())
}

func TestInValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	res := h(input.ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
}
