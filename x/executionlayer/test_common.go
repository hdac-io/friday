package executionlayer

import (
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/store"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	authtypes "github.com/hdac-io/friday/x/auth/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/hdac-io/friday/x/params/subspace"
	"github.com/hdac-io/friday/x/readablename"
	abci "github.com/hdac-io/tendermint/abci/types"
	"github.com/hdac-io/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const (
	chainID = "test-chain-id"
)

var (
	GenesisAccountAddress, _   = sdk.AccAddressFromBech32("friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz")
	GenesisPubKeyString        = "fridaypub1addwnpepqw6vr6728nvg2duwj062y2yx2mfhmqjh66mjtgsyf7jwyq2kx2kaqlkq94l"
	GenesisPubKey              = sdk.MustGetAccPubKeyBech32(GenesisPubKeyString)
	RecipientAccountAddress, _ = sdk.AccAddressFromBech32("friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv")
	RecipientPubKeyString      = "fridaypub1addwnpepqg3xvg45h4j0wsj6dng0wcze2vwvnc7hse696xjvy0cwk347zm0lvhcskq7"
	RecipientPubKey            = sdk.MustGetAccPubKeyBech32(RecipientPubKeyString)

	contractPath        = os.ExpandEnv("$HOME/.nodef/contracts")
	mintInstallWasm     = "mint_install.wasm"
	posInstallWasm      = "pos_install.wasm"
	standardPaymentWasm = "standard_payment.wasm"
	counterDefineWasm   = "counter_define.wasm"
	counterCallWasm     = "counter_call.wasm"
	transferWasm        = "transfer_to_account.wasm"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	elk ExecutionLayerKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	hashMapStoreKey := sdk.NewKVStoreKey(HashMapStoreKey)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	readablenameStoreKey := sdk.NewKVStoreKey("readablename")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, authtypes.DefaultParamspace)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(readablenameStoreKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(hashMapStoreKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: chainID}, false, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authCapKey, ps, auth.ProtoBaseAccount)
	readablenameKeeper := readablename.NewReadableNameKeeper(readablenameStoreKey, cdc)

	elk := NewExecutionLayerKeeper(cdc, hashMapStoreKey, os.ExpandEnv("$HOME/.casperlabs/.casper-node.sock"),
		accountKeeper, readablenameKeeper)

	gs := types.DefaultGenesisState()
	gs.ChainName = chainID
	gs.Accounts = make([]types.Account, 1)
	pubkey := sdk.MustGetSecp256k1FromBech32AccPubKey(GenesisPubKeyString)
	gs.Accounts[0] = types.Account{
		PublicKey:           *pubkey,
		InitialBalance:      "500000000",
		InitialBondedAmount: "1000000",
	}
	elk.SetGenesisConf(ctx, gs.GenesisConf)
	elk.SetGenesisAccounts(ctx, gs.Accounts)

	return testInput{
		cdc: cdc,
		ctx: ctx,
		elk: elk,
	}
}

func genesis(keeper ExecutionLayerKeeper) []byte {
	input := setupTestInput()
	genesisState := types.GenesisState{
		GenesisConf: input.elk.GetGenesisConf(input.ctx),
		Accounts:    input.elk.GetGenesisAccounts(input.ctx),
	}
	genesisConfig, err := types.ToChainSpecGenesisConfig(genesisState)
	if err != nil {
		panic(err)
	}
	response, err := grpc.RunGenesis(keeper.client, genesisConfig)

	if err != nil {
		panic(err)
	}

	return response.GetSuccess().PoststateHash
}

func counterDefine(keeper ExecutionLayerKeeper, parentStateHash []byte) []byte {
	input := setupTestInput()
	timestamp := time.Now().Unix()
	paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000))
	cntDefCode := util.LoadWasmFile(path.Join(contractPath, counterDefineWasm))
	standardPaymentCode := util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm))
	protocolVersion := input.elk.MustGetProtocolVersion(input.ctx)
	genesisAddr := sdk.MustGetEEAddressFromCryptoPubkey(input.elk.GetGenesisAccounts(input.ctx)[0].PublicKey)

	deploy := util.MakeDeploy(genesisAddr.Bytes(), cntDefCode, []byte{},
		standardPaymentCode, paymentAbi, uint64(10), timestamp, input.elk.GetChainName(input.ctx))

	deploys := util.MakeInitDeploys()
	deploys = util.AddDeploy(deploys, deploy)

	effects2, grpcErr := grpc.Execute(keeper.client, parentStateHash, timestamp, deploys, &protocolVersion)
	if grpcErr != "" {
		panic(fmt.Sprintf("counter define execute: %s", grpcErr))
	}

	postStateHash, _, grpcErr := grpc.Commit(keeper.client, parentStateHash, effects2, &protocolVersion)
	if grpcErr != "" {
		panic(fmt.Sprintf("counter define commmit: %s", grpcErr))
	}

	return postStateHash

}

func counterCall(keeper ExecutionLayerKeeper, parentStateHash []byte) []byte {
	input := setupTestInput()
	timestamp := time.Now().Unix()
	paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000))
	cntCallCode := util.LoadWasmFile(path.Join(contractPath, counterCallWasm))
	standardPaymentCode := util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm))
	protocolVersion := input.elk.MustGetProtocolVersion(input.ctx)
	genesisAddr := sdk.MustGetEEAddressFromCryptoPubkey(input.elk.GetGenesisAccounts(input.ctx)[0].PublicKey)

	timestamp = time.Now().Unix()
	deploy := util.MakeDeploy(genesisAddr.Bytes(), cntCallCode,
		[]byte{}, standardPaymentCode, paymentAbi, uint64(10), timestamp, input.elk.GetChainName(input.ctx))
	deploys := util.MakeInitDeploys()
	deploys = util.AddDeploy(deploys, deploy)

	effects3, grpcErr := grpc.Execute(keeper.client, parentStateHash, timestamp, deploys, &protocolVersion)
	if grpcErr != "" {
		panic(fmt.Sprintf("counter call execute: %s", grpcErr))
	}

	postStateHash, _, grpcErr := grpc.Commit(keeper.client, parentStateHash, effects3, &protocolVersion)
	if grpcErr != "" {
		panic(fmt.Sprintf("counter call commit: %s", grpcErr))
	}

	return postStateHash
}
