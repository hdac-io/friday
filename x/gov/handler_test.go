package gov

import (
	"strings"
	"testing"

	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"

	"github.com/stretchr/testify/require"
)

func TestInvalidMsg(t *testing.T) {
	k := Keeper{}
	h := NewHandler(k)

	res := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg(), false, 0, 0)
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized gov message type"))
}
