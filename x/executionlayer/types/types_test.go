package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPublicKey(t *testing.T) {
	// valid input
	ed25519PublicKey := "lGBOOzBXMuDEPxDIE5of4u9U+yzmnrB2MbjA0dQM9LQ="
	publicKey, err := NewPublicKey(ed25519PublicKey)
	require.Nil(t, err)
	require.NotNil(t, publicKey)

	// invalid input.
	// base64 encoded but not 32byte length
	invalidPublicKey := "YXNkZmdoamtsO3F3ZXJ0eXVpb3A="
	publicKey, err = NewPublicKey(invalidPublicKey)
	require.NotNil(t, err)
	require.Nil(t, publicKey)

	// 32byte length but not base64 encoded
	invalidPublicKey = "12345678901234567890123456789012"
	publicKey, err = NewPublicKey(invalidPublicKey)
	require.NotNil(t, err)
	require.Nil(t, publicKey)
}
