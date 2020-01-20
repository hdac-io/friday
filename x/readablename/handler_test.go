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

	cryptoPubkey := secp256k1.GenPrivKey().PubKey()
	var pubkey secp256k1.PubKeySecp256k1
	input.cdc.MustUnmarshalBinaryBare(cryptoPubkey.Bytes(), &pubkey)

	addr := sdk.AccAddress(cryptoPubkey.Address())
	res := h(input.ctx, NewMsgSetAccount(NewName("bryanrhee"), addr, pubkey))
	require.True(t, res.IsOK())

	newCryptopubkey := secp256k1.GenPrivKey().PubKey()
	var newpubkey secp256k1.PubKeySecp256k1
	input.cdc.MustUnmarshalBinaryBare(newCryptopubkey.Bytes(), &pubkey)
	newaddr := sdk.AccAddress(newCryptopubkey.Address())
	res = h(input.ctx, NewMsgChangeKey("bryanrhee", addr, newaddr, pubkey, newpubkey))
	require.True(t, res.IsOK())
}

func TestInValidMsg(t *testing.T) {
	input := setupTestInput()
	h := NewHandler(input.k)

	res := h(input.ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
}
