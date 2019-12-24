package executionlayer

import (
	"fmt"
	"math/big"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/x/executionlayer/types"
	abci "github.com/hdac-io/tendermint/abci/types"
	"github.com/stretchr/testify/assert"
)

//-------------------------------------------

func TestGetQueryResult(t *testing.T) {
	input := setupTestInput()
	queryPath := "counter/count"

	parentHash := genesis(input.elk)

	parentHash = counterDefine(input.elk, parentHash)
	parentHash = counterCall(input.elk, parentHash)

	res, err := input.elk.GetQueryResult(
		input.ctx,
		parentHash,
		"address", input.genesisAddress.String(), queryPath)

	if err != nil {
		fmt.Println(err.Error())
		panic("Fail to execute")
	}

	fmt.Println(res)

	assert.NotNil(t, res)
}

func TestGetQueryBalanceResult(t *testing.T) {
	input := setupTestInput()
	parentHash := genesis(input.elk)
	res, err := input.elk.GetQueryBalanceResult(input.ctx, parentHash, types.ToPublicKey(input.genesisAddress))

	if err != nil {
		fmt.Println(err.Error())
		panic("Fail to execute")
	}

	fmt.Println(res)

	assert.NotNil(t, res)
}

func TestUnitHashMapNormalInput(t *testing.T) {
	input := setupTestInput()

	blockHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	eeState := []byte{31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	result := input.elk.SetEEState(input.ctx, blockHash, eeState)
	assert.Equal(t, true, result)

	unitHash := input.elk.GetUnitHashMap(input.ctx, blockHash)
	assert.Equal(t, eeState, unitHash.EEState)
}

func TestUnitHashMapInCorrectInput(t *testing.T) {
	input := setupTestInput()

	blockHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	eeState := []byte{31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	result := input.elk.SetEEState(input.ctx, blockHash, eeState)
	assert.Equal(t, false, result)

	unitHash := input.elk.GetUnitHashMap(input.ctx, blockHash)
	assert.NotEqual(t, eeState, unitHash.EEState)
}

func TestCreateBlock(t *testing.T) {
	input := setupTestInput()
	parentHash := genesis(input.elk)
	input.elk.SetEEState(input.ctx, input.blockHash, parentHash)
	queryPath := "counter/count"

	blockHash1 := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	blockHash2 := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	counterDefineMSG := NewMsgExecute(
		input.blockHash,
		input.genesisAddress,
		input.genesisAddress,
		util.LoadWasmFile(path.Join(contractPath, counterDefineWasm)),
		[]byte{},
		util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm)),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(10),
	)

	handlerMsgExecute(input.ctx, input.elk, counterDefineMSG)

	nextBlockABCI1 := abci.RequestBeginBlock{
		Hash:   blockHash1,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: input.blockHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI1, input.elk)

	counterCallMSG := NewMsgExecute(
		input.blockHash,
		input.genesisAddress,
		input.genesisAddress,
		util.LoadWasmFile(path.Join(contractPath, counterCallWasm)),
		[]byte{},
		util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm)),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(10),
	)

	handlerMsgExecute(input.ctx, input.elk, counterCallMSG)

	nextBlockABCI2 := abci.RequestBeginBlock{
		Hash:   blockHash2,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: input.blockHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI2, input.elk)

	arrPath := strings.Split(queryPath, "/")

	unitHash1 := input.elk.GetUnitHashMap(input.ctx, blockHash1)
	res1, _ := grpc.Query(input.elk.client, unitHash1.EEState, "address", types.ToPublicKey(input.genesisAddress), arrPath, input.protocolVersion)
	assert.Equal(t, int32(0), res1.GetIntValue())

	unitHash2 := input.elk.GetUnitHashMap(input.ctx, blockHash2)
	res2, _ := grpc.Query(input.elk.client, unitHash2.EEState, "address", types.ToPublicKey(input.genesisAddress), arrPath, input.protocolVersion)
	assert.Equal(t, int32(1), res2.GetIntValue())
}

func TestMarsahlAndUnMarshal(t *testing.T) {
	src := &transforms.TransformEntry{
		Transform: &transforms.Transform{TransformInstance: &transforms.Transform_Write{Write: &transforms.TransformWrite{Value: &state.Value{Value: &state.Value_IntValue{IntValue: 1}}}}}}
	bz, _ := proto.Marshal(src)

	obj := &transforms.TransformEntry{}
	proto.Unmarshal(bz, obj)

	assert.Equal(t, src.Transform.String(), obj.Transform.String())
}

func TestGenesisState(t *testing.T) {
	testMock := setupTestInput()

	expected := types.DefaultGenesisState()
	testMock.elk.SetGenesisConf(testMock.ctx, expected.GenesisConf)
	testMock.elk.SetGenesisAccounts(testMock.ctx, expected.Accounts)

	var got types.GenesisState
	got.GenesisConf = testMock.elk.GetGenesisConf(testMock.ctx)
	got.Accounts = testMock.elk.GetGenesisAccounts(testMock.ctx)
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %v, but got: %v", expected, got)
	}

	// accounts Marshal, UnMarshal test
	expected.Accounts = make([]types.Account, 1)
	expected.Accounts[0].PublicKey = types.PublicKey([]byte("test-pub-key"))
	expected.Accounts[0].InitialBalance = "2"
	expected.Accounts[0].InitialBondedAmount = "1"

	testMock.elk.SetGenesisAccounts(testMock.ctx, expected.Accounts)
	gottonAccounts := testMock.elk.GetGenesisAccounts(testMock.ctx)
	if !reflect.DeepEqual(expected.Accounts, gottonAccounts) {
		t.Errorf("expected: %v, but got: %v", expected.Accounts, gottonAccounts)
	}
}
