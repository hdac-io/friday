package nameservice

import (
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/store"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/params"
	abci "github.com/hdac-io/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
	"github.com/hdac-io/tendermint/libs/log"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	k   AccountKeeper
	ak  auth.AccountKeeper
	pk  params.Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterCodec(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	fckCapKey := sdk.NewKVStoreKey("fckCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")
	storekey := sdk.NewKVStoreKey(StoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(fckCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(storekey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(
		cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, auth.DefaultParams())
	storeKeeper := NewAccountKeeper(storekey, cdc)

	return testInput{cdc: cdc, ctx: ctx, k: storeKeeper, ak: ak, pk: pk}
}
