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
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

var (
	contractPath        = os.ExpandEnv("$HOME/.nodef/contracts")
	mintInstallWasm     = "mint_install.wasm"
	posInstallWasm      = "pos_install.wasm"
	standardPaymentWasm = "standard_payment.wasm"
	counterDefineWasm   = "counter_define.wasm"
	counterCallWasm     = "counter_call.wasm"
)

type testInput struct {
	cdc               *codec.Codec
	ctx               sdk.Context
	elk               ExecutionLayerKeeper
	blockHash         []byte
	genesisAddress    sdk.AccAddress
	strGenesisAddress string
	genesisAccount    map[string][]string
	chainName         string
	costs             map[string]uint32
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

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	blockHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	strGenesisAddress := "friday1dl2cjlfpmc9hcyd4rxts047tze87s0gxmzqx70"
	genesisAddress, _ := sdk.AccAddressFromBech32(strGenesisAddress)
	chainName := "hdac"
	accounts := map[string][]string{
		strGenesisAddress: []string{"500000000", "1000000"}}

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

	elk := NewExecutionLayerKeeper(cdc, hashMapStoreKey, os.ExpandEnv("$HOME/.casperlabs/.casper-node.sock"), "1.0.0")

	elk.InitialUnitHashMap(ctx, blockHash)

	return testInput{
		cdc:               cdc,
		ctx:               ctx,
		elk:               elk,
		blockHash:         blockHash,
		genesisAddress:    genesisAddress,
		strGenesisAddress: strGenesisAddress,
		genesisAccount:    accounts,
		chainName:         chainName,
		costs:             costs,
	}
}

func genesis(keeper ExecutionLayerKeeper) []byte {
	input := setupTestInput()
	fmt.Printf("%v", input.genesisAccount)
	genesisConfig, err := util.GenesisConfigMock(
		input.chainName, types.ToPublicKey(input.genesisAddress),
		input.genesisAccount[input.strGenesisAddress][0],
		input.genesisAccount[input.strGenesisAddress][1],
		input.elk.protocolVersion, input.costs, path.Join(contractPath, mintInstallWasm),
		path.Join(contractPath, posInstallWasm),
	)

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

	deploy := util.MakeDeploy(types.ToPublicKey(input.genesisAddress), cntDefCode, []byte{},
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
	cntCallCode := util.LoadWasmFile(path.Join(contractPath, counterCallWasm))
	standardPaymentCode := util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm))

	timestamp = time.Now().Unix()
	deploy := util.MakeDeploy(types.ToPublicKey(input.genesisAddress), cntCallCode,
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
