package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/hdac-io/friday/types"
)

func TestNewPublicKey(t *testing.T) {
	// valid input
	bech32PublicKey := "fridaypub1addwnpepqw6vr6728nvg2duwj062y2yx2mfhmqjh66mjtgsyf7jwyq2kx2kaqlkq94l"
	publicKeyFromBech32, err := NewPublicKey(bech32PublicKey)
	require.Nil(t, err)
	require.NotNil(t, publicKeyFromBech32)

	// invalid input.
	// base64 encoded but not 32byte length
	invalidPublicKey := "YXNkZmdoamtsO3F3ZXJ0eXVpb3A="
	publicKey, err := NewPublicKey(invalidPublicKey)
	require.NotNil(t, err)
	require.Nil(t, publicKey)

	// 32byte length but not base64 encoded
	invalidPublicKey = "12345678901234567890123456789012"
	publicKey, err = NewPublicKey(invalidPublicKey)
	require.NotNil(t, err)
	require.Nil(t, publicKey)

	cryptoPublicKey := sdk.MustGetAccPubKeyBech32(bech32PublicKey)
	publicKeyFromCrypto := NewPublicKeyFromCryptoPubkey(cryptoPublicKey)
	require.Equal(t, publicKeyFromCrypto, publicKeyFromBech32)
}
