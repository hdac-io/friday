package nameservice

import (
	"testing"

	sdk "github.com/hdac-io/friday/types"

	"github.com/stretchr/testify/require"
)

func TestValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	addr, _ := sdk.AccAddressFromBech32("hdac-io1deadn2dxuy4ls6x7p2mw9prvw3nfhfvph974zt")
	res := h(input.ctx, NewMsgSetAccount("bryanrhee", addr))
	require.True(t, res.IsOK())

	res = h(input.ctx, NewMsgAddrCheck("bryanrhee", addr))
	require.True(t, res.IsOK())

	res = h(input.ctx, NewMsgChangeKey("bryanrhee", addr, addr))
	require.True(t, res.IsOK())
}

func TestInValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	res := h(input.ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
}
