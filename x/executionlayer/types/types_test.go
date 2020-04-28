package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hdac-io/friday/types"
)

func TestNewPublicKeyFromAddress(t *testing.T) {
	// valid input
	bech32ValAddr := "fridayvaloper1gp2u22697kz6slwa25k2tkhz6st2l0zx3hkfc5wdlpjaauv5czsqcd3mya"
	byteAddr, err := sdk.GetFromBech32(bech32ValAddr, "fridayvaloper")
	require.Nil(t, err)

	valAddr := sdk.ValAddress(byteAddr)
	require.Equal(t, len(valAddr), 32)
}
