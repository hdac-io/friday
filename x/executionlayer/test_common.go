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
	"github.com/hdac-io/friday/x/executionlayer/types"
	abci "github.com/hdac-io/tendermint/abci/types"
	"github.com/hdac-io/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const (
	chainID = "test-chain-id"
)

var (
	GenesisAccountAddress, _ = sdk.AccAddressFromBech32("friday1dl2cjlfpmc9hcyd4rxts047tze87s0gxmzqx70")
	contractPath             = os.ExpandEnv("$HOME/.nodef/contracts")
	mintInstallWasm          = "mint_install.wasm"
	posInstallWasm           = "pos_install.wasm"
	standardPaymentWasm      = "standard_payment.wasm"
	counterDefineWasm        = "counter_define.wasm"
	counterCallWasm          = "counter_call.wasm"
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

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(hashMapStoreKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: chainID}, false, log.NewNopLogger())

	elk := NewExecutionLayerKeeper(cdc, hashMapStoreKey, os.ExpandEnv("$HOME/.casperlabs/.casper-node.sock"))

	gs := types.DefaultGenesisState()
	gs.GenesisConf.Genesis.Name = chainID
	gs.Accounts = make([]types.Account, 1)
	gs.Accounts[0] = types.Account{
		PublicKey:           types.ToPublicKey(GenesisAccountAddress),
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

	deploy := util.MakeDeploy(input.elk.GetGenesisAccounts(input.ctx)[0].PublicKey, cntDefCode, []byte{},
		standardPaymentCode, paymentAbi, uint64(10), timestamp, input.elk.GetGenesisConf(input.ctx).Genesis.Name)

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

	timestamp = time.Now().Unix()
	deploy := util.MakeDeploy(input.elk.GetGenesisAccounts(input.ctx)[0].PublicKey, cntCallCode,

		[]byte{}, standardPaymentCode, paymentAbi, uint64(10), timestamp, input.elk.GetGenesisConf(input.ctx).Genesis.Name)
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
