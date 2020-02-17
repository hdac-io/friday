package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hdac-io/friday/types"
)

func TestNewPublicKeyFromAddress(t *testing.T) {
	// valid input
	bech32ValAddr := "fridayvaloper15evpva2u57vv6l5czehyk69s0wnq9hrk4gqxv2"
	byteAddr, err := sdk.GetFromBech32(bech32ValAddr, "fridayvaloper")
	require.Nil(t, err)

	valAddr := sdk.ValAddress(byteAddr)
	require.Equal(t, len(valAddr.ToEEAddress()), 32)
}
