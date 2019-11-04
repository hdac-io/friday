package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hdac-io/friday/types"
)

func TestMsgUnjailGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("abcd")
	msg := NewMsgUnjail(sdk.ValAddress(addr))
	bytes := msg.GetSignBytes()
	require.Equal(
		t,
		`{"type":"friday/MsgUnjail","value":{"address":"fridayvaloper1v93xxeqhg9nn6"}}`,
		string(bytes),
	)
}
