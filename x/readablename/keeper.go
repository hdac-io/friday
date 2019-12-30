package readablename

import (
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/tendermint/crypto"

	sdk "github.com/hdac-io/friday/types"
)

/*
Requirements:
	- Provide account service with readable ID

Impl:
	- [Readable ID : Public key] matching logic
	- Key checking logic for account login
	- Key change logic
	- Duplicate readable ID check

*/

// ReadableNameKeeper is a store of all the account we've seen,
// and amino codec
// Will further check more why store struct should contain codec
type ReadableNameKeeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
}

// NewReadableNameKeeper returns AccountStore DB object
func NewReadableNameKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) ReadableNameKeeper {
	return ReadableNameKeeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// GetUnitAccount fetches the AccountInfo with the given unit account data
// If not found, acc.UnitAccount is nil.
func (k *ReadableNameKeeper) GetUnitAccount(ctx sdk.Context, name string) UnitAccount {
	st := ctx.KVStore(k.storeKey)
	if !st.Has([]byte(name)) {
		return UnitAccount{}
	}
	val := st.Get([]byte(name))
	var acc UnitAccount
	k.cdc.MustUnmarshalBinaryBare(val, &acc)
	return acc
}

// SetUnitAccount adds the given unit account to the database.
// It returns false if the account is already stored.
func (k *ReadableNameKeeper) SetUnitAccount(ctx sdk.Context, name string, address sdk.AccAddress, pubkey crypto.PubKey) bool {
	// check if we already have seen it
	acc := k.GetUnitAccount(ctx, name)
	if acc.Name.MustToString() != "" {
		return false
	}

	// Constructring & Marshal
	acc = NewUnitAccount(NewName(name), address, pubkey)
	accBytes := k.cdc.MustMarshalBinaryBare(acc)

	// add it to the store
	st := ctx.KVStore(k.storeKey)
	st.Set([]byte(name), accBytes)

	return true
}

// ChangeKey updates public key of the account and apply to the database
func (k *ReadableNameKeeper) ChangeKey(ctx sdk.Context, name string,
	oldAddr, newAddr sdk.AccAddress,
	oldpubkey, newpubkey crypto.PubKey) bool {

	// check if we already have seen it
	acc := k.GetUnitAccount(ctx, name)
	if acc.Address.String() != oldAddr.String() {
		return false
	}

	acc = NewUnitAccount(NewName(name), newAddr, newpubkey)
	accBytes := k.cdc.MustMarshalBinaryBare(acc)

	// add it to the store
	st := ctx.KVStore(k.storeKey)
	st.Set([]byte(name), accBytes)

	return true
}

// AddrCheck checks account by given address
func (k *ReadableNameKeeper) AddrCheck(ctx sdk.Context, name string, address sdk.AccAddress) bool {
	acc := k.GetUnitAccount(ctx, name)
	strName := acc.Name.MustToString()
	if acc.Address.String() == address.String() && strName != "" {
		return true
	}
	return false
}

// PubKeyCheck checks account by given public key
func (k *ReadableNameKeeper) PubKeyCheck(ctx sdk.Context, name string, pubkey crypto.PubKey) bool {
	acc := k.GetUnitAccount(ctx, name)
	strName := acc.Name.MustToString()
	if acc.PubKey.Equals(pubkey) && strName != "" {
		return true
	}
	return false
}

// GetAccountIterator get iterator for listting all accounts.
func (k *ReadableNameKeeper) GetAccountIterator(ctx sdk.Context) sdk.Iterator {
	str := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(str, nil)
}
