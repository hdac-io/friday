package nickname

import (
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/x/auth"

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

// NicknameKeeper is a store of all the account we've seen,
// and amino codec
// Will further check more why store struct should contain codec
type NicknameKeeper struct {
	cdc           *codec.Codec
	storeKey      sdk.StoreKey
	AccountKeeper auth.AccountKeeper
}

// NewNicknameKeeper returns AccountStore DB object
func NewNicknameKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, k auth.AccountKeeper) NicknameKeeper {
	return NicknameKeeper{
		storeKey:      storeKey,
		cdc:           cdc,
		AccountKeeper: k,
	}
}

// GetUnitAccount fetches the AccountInfo with the given unit account data
// If not found, acc.UnitAccount is nil.
func (k *NicknameKeeper) GetUnitAccount(ctx sdk.Context, name string) UnitAccount {
	st := ctx.KVStore(k.storeKey)
	if !st.Has([]byte(name)) {
		return UnitAccount{}
	}
	val := st.Get([]byte(name))
	var acc UnitAccount
	k.cdc.MustUnmarshalBinaryBare(val, &acc)
	return acc
}

// SetNickname adds the given unit account to the database.
// It returns false if the account is already stored.
func (k *NicknameKeeper) SetNickname(ctx sdk.Context, name string, address sdk.AccAddress) bool {
	// check if we already have seen it
	acc := k.GetUnitAccount(ctx, name)
	if acc.Nickname.MustToString() != "" {
		return false
	}

	// Constructring & Marshal
	acc = NewUnitAccount(NewName(name), address)
	accBytes := k.cdc.MustMarshalBinaryBare(acc)

	// add it to the store
	st := ctx.KVStore(k.storeKey)
	st.Set([]byte(name), accBytes)

	return true
}

// ChangeKey updates public key of the account and apply to the database
func (k *NicknameKeeper) ChangeKey(ctx sdk.Context, name string, oldAddr, newAddr sdk.AccAddress) bool {

	// check if we already have seen it
	acc := k.GetUnitAccount(ctx, name)
	if acc.Address.String() != oldAddr.String() {
		return false
	}

	k.SetAccountIfNotExists(ctx, newAddr)
	acc = NewUnitAccount(NewName(name), newAddr)
	accBytes := k.cdc.MustMarshalBinaryBare(acc)

	// add it to the store
	st := ctx.KVStore(k.storeKey)
	st.Set([]byte(name), accBytes)

	return true
}

// AddrCheck checks account by given address
func (k *NicknameKeeper) AddrCheck(ctx sdk.Context, name string, address sdk.AccAddress) bool {
	acc := k.GetUnitAccount(ctx, name)
	strName := acc.Nickname.MustToString()
	if acc.Address.String() == address.String() && strName != "" {
		return true
	}
	return false
}

// GetAccountIterator get iterator for listting all accounts.
func (k *NicknameKeeper) GetAccountIterator(ctx sdk.Context) sdk.Iterator {
	str := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(str, nil)
}

// SetAccountIfNotExists runs if network has no given account
func (k NicknameKeeper) SetAccountIfNotExists(ctx sdk.Context, addr sdk.AccAddress) {
	// Recepient account existence check, if not, create one
	toAddressAccountObject := k.AccountKeeper.GetAccount(ctx, addr)
	if toAddressAccountObject == nil {
		toAddressAccountObject = k.AccountKeeper.NewAccountWithAddress(ctx, addr)
		k.AccountKeeper.SetAccount(ctx, toAddressAccountObject)
	}
}
