package nickname

import (
	"testing"

	sdk "github.com/hdac-io/friday/types"

	"github.com/hdac-io/tendermint/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	res := h(input.ctx, NewMsgSetAccount(NewName("bryanrhee"), addr), false, 0, 0)
	require.True(t, res.IsOK())

	newaddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	res = h(input.ctx, NewMsgChangeKey("bryanrhee", addr, newaddr), false, 0, 0)
	require.True(t, res.IsOK())
}

func TestInValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	res := h(input.ctx, sdk.NewTestMsg(), false, 0, 0)
	require.False(t, res.IsOK())
}
