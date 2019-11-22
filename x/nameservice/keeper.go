package nameservice

import (
	"github.com/hdac-io/friday/codec"

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

// AccountKeeper is a store of all the account we've seen,
// and amino codec
// Will further check more why store struct should contain codec
type AccountKeeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
}

// NewAccountKeeper returns AccountStore DB object
func NewAccountKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) AccountKeeper {
	return AccountKeeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// GetUnitAccount fetches the AccountInfo with the given unit account data
// If not found, acc.UnitAccount is nil.
func (k *AccountKeeper) GetUnitAccount(ctx sdk.Context, name string) UnitAccount {
	st := ctx.KVStore(k.storeKey)
	if !st.Has([]byte(name)) {
		return NewUnitAccount()
	}
	val := st.Get([]byte(name))
	var acc UnitAccount
	k.cdc.MustUnmarshalBinaryBare(val, &acc)
	return acc
}

// SetUnitAccount adds the given unit account to the database.
// It returns false if the account is already stored.
func (k *AccountKeeper) SetUnitAccount(ctx sdk.Context, name string, address sdk.AccAddress) bool {
	// check if we already have seen it
	acc := k.GetUnitAccount(ctx, name)
	strname, _ := acc.ID.ToString()
	if strname != "" {
		return false
	}

	// Constructring & Marshal
	acc = UnitAccount{
		ID:      NewName(name),
		Address: address,
	}
	accBytes := k.cdc.MustMarshalBinaryBare(acc)

	// add it to the store
	st := ctx.KVStore(k.storeKey)
	st.Set([]byte(name), accBytes)

	return true
}

// ChangeKey updates public key of the account and apply to the database
func (k *AccountKeeper) ChangeKey(ctx sdk.Context, name string, oldAddr, newAddr sdk.AccAddress) bool {
	// check if we already have seen it
	acc := k.GetUnitAccount(ctx, name)
	if acc.Address.String() != oldAddr.String() {
		return false
	}

	acc = UnitAccount{
		ID:      NewName(name),
		Address: newAddr,
	}
	accBytes := k.cdc.MustMarshalBinaryBare(acc)

	// add it to the store
	st := ctx.KVStore(k.storeKey)
	st.Set([]byte(name), accBytes)

	return true
}

// AddrCheck checks account by given private key
func (k *AccountKeeper) AddrCheck(ctx sdk.Context, name string, address sdk.AccAddress) bool {
	acc := k.GetUnitAccount(ctx, name)
	strName, _ := acc.ID.ToString()
	if acc.Address.String() == address.String() && strName != "" {
		return true
	}
	return false
}

// GetAccountIterator get iterator for listting all accounts.
func (k *AccountKeeper) GetAccountIterator(ctx sdk.Context) sdk.Iterator {
	str := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(str, nil)
}
