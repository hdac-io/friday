package executionlayer

import (
	"fmt"
	"math/big"
	"time"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/store"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testInput struct {
	cdc            *codec.Codec
	ctx            sdk.Context
	rootState      []byte
	genesisAddress string
	genesisAccount map[string][]string
	chainName      string
	costs          map[string]uint32
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	unitHashMapKey := sdk.NewKVStoreKey("unitHashMapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(unitHashMapKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	emptyStateHash := util.DecodeHexString(util.StrEmptyStateHash)
	genesisAddress := "d70243dd9d0d646fd6df282a8f7a8fa05a6629bec01d8024c3611eb1c1fb9f84"
	chainName := "hdac"
	accounts := map[string][]string{
		genesisAddress: []string{"500000000", "1000000"}}

	costs := map[string]uint32{
		"regular":            1,
		"div-multiplier":     16,
		"mul-multiplier":     4,
		"mem-multiplier":     2,
		"mem-initial-pages":  4096,
		"mem-grow-per-page":  8192,
		"mem-copy-per-byte":  1,
		"max-stack-height":   65536,
		"opcodes-multiplier": 3,
		"opcodes-divisor":    8}

	return testInput{
		cdc:            cdc,
		ctx:            ctx,
		rootState:      emptyStateHash,
		genesisAddress: genesisAddress,
		genesisAccount: accounts,
		chainName:      chainName,
		costs:          costs,
	}
}

func genesis(keeper ExecutionLayerKeeper) []byte {
	input := setupTestInput()
	mintCode := util.LoadWasmFile("./wasms/mint_install.wasm")
	posCode := util.LoadWasmFile("./wasms/pos_install.wasm")

	_, effects, errMsg := grpc.RunGenesis(keeper.client,
		input.chainName,
		0,
		keeper.protocolVersion,
		mintCode,
		posCode,
		input.genesisAccount,
		input.costs)

	if errMsg != "" {
		panic(errMsg)
	}

	postStateHash, _, errMessage := grpc.Commit(keeper.client, input.rootState, effects, keeper.protocolVersion)

	if errMessage != "" {
		panic(errMessage)
	}

	return postStateHash
}

func counterDefine(keeper ExecutionLayerKeeper, parentStateHash []byte) []byte {
	input := setupTestInput()
	timestamp := time.Now().Unix()
	paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000))
	cntDefCode := util.LoadWasmFile("./wasms/counter_define.wasm")
	standardPaymentCode := util.LoadWasmFile("./wasms/standard_payment.wasm")

	deploy := util.MakeDeploy(input.genesisAddress, cntDefCode, []byte{},
		standardPaymentCode, paymentAbi, uint64(10), timestamp, input.chainName)

	deploys := util.MakeInitDeploys()
	deploys = util.AddDeploy(deploys, deploy)

	effects2, err := grpc.Execute(keeper.client, parentStateHash, timestamp, deploys, keeper.protocolVersion)
	if err != "" {
		panic(fmt.Sprintf("counter define execute: %s", err))
	}

	postStateHash, _, err := grpc.Commit(keeper.client, parentStateHash, effects2, keeper.protocolVersion)
	if err != "" {
		panic(fmt.Sprintf("counter define commmit: %s", err))
	}

	return postStateHash

}

func counterCall(keeper ExecutionLayerKeeper, parentStateHash []byte) []byte {
	input := setupTestInput()
	timestamp := time.Now().Unix()
	paymentAbi := util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000))
	cntCallCode := util.LoadWasmFile("./wasms/counter_call.wasm")
	standardPaymentCode := util.LoadWasmFile("./wasms/standard_payment.wasm")

	timestamp = time.Now().Unix()
	deploy := util.MakeDeploy(input.genesisAddress, cntCallCode,
		[]byte{}, standardPaymentCode, paymentAbi, uint64(10), timestamp, input.chainName)
	deploys := util.MakeInitDeploys()
	deploys = util.AddDeploy(deploys, deploy)

	effects3, err := grpc.Execute(keeper.client, parentStateHash, timestamp, deploys, keeper.protocolVersion)
	if err != "" {
		panic(fmt.Sprintf("counter call execute: %s", err))
	}

	postStateHash, _, err := grpc.Commit(keeper.client, parentStateHash, effects3, keeper.protocolVersion)
	if err != "" {
		panic(fmt.Sprintf("counter call commit: %s", err))
	}

	return postStateHash
}
