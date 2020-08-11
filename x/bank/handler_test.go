package bank

import (
	"strings"
	"testing"

	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"

	"github.com/stretchr/testify/require"
)

func TestInvalidMsg(t *testing.T) {
	h := NewHandler(nil)

	res := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg(), false, 0, 0)
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized bank message type"))
}
