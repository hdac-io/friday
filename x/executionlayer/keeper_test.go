package executionlayer

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

//-------------------------------------------

func TestGetQueryResult(t *testing.T) {
	input := setupTestInput()
	path := "counter/count"

	parentHash := genesis(input.elk)

	parentHash = counterDefine(input.elk, parentHash)
	parentHash = counterCall(input.elk, parentHash)

	res, err := input.elk.GetQueryResult(
		input.ctx,
		parentHash,
		"address", input.genesisAddress, path)

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
	res, err := input.elk.GetQueryBalanceResult(input.ctx, parentHash, input.genesisAddress)

	if err != nil {
		fmt.Println(err.Error())
		panic("Fail to execute")
	}

	fmt.Println(res)

	assert.NotNil(t, res)
}

func TestUnitHashMapNormalInput(t *testing.T) {
	input := setupTestInput()

	blockState := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	eeState := []byte{31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	result := input.elk.SetUnitHashMap(input.ctx, blockState, eeState)
	assert.Equal(t, true, result)

	resEEState := input.elk.GetEEState(input.ctx, blockState)
	resNextEEState := input.elk.GetNextState(input.ctx, blockState)
	assert.Equal(t, eeState, resEEState)
	assert.Equal(t, eeState, resNextEEState)

	nextState := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	input.elk.SetNextState(input.ctx, blockState, nextState)
	resEEState = input.elk.GetEEState(input.ctx, blockState)
	resNextEEState = input.elk.GetNextState(input.ctx, blockState)
	assert.Equal(t, eeState, resEEState)
	assert.Equal(t, nextState, resNextEEState)
}

func TestUnitHashMapInCorrectInput(t *testing.T) {
	input := setupTestInput()

	blockState := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	eeState := []byte{31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	result := input.elk.SetUnitHashMap(input.ctx, blockState, eeState)
	assert.Equal(t, false, result)

	res := input.elk.GetEEState(input.ctx, blockState)
	assert.NotEqual(t, eeState, res)
}

func TestCreateBlock(t *testing.T) {
	input := setupTestInput()
	parentHash := genesis(input.elk)
	input.elk.SetNextState(input.ctx, input.blockStateHash, parentHash)
	path := "counter/count"

	blockState1 := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	blockState2 := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	counterDefineMSG := NewMsgExecute(
		input.blockStateHash,
		util.DecodeHexString(input.genesisAddress),
		util.DecodeHexString(input.genesisAddress),
		util.LoadWasmFile("./wasms/counter_define.wasm"),
		[]byte{},
		util.LoadWasmFile("./wasms/standard_payment.wasm"),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(10),
	)

	handlerMsgExecute(input.ctx, input.elk, counterDefineMSG)

	nextBlockABCI1 := abci.RequestBeginBlock{
		Hash:   blockState1,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: input.blockStateHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI1, input.elk)

	counterCallMSG := NewMsgExecute(
		input.blockStateHash,
		util.DecodeHexString(input.genesisAddress),
		util.DecodeHexString(input.genesisAddress),
		util.LoadWasmFile("./wasms/counter_call.wasm"),
		[]byte{},
		util.LoadWasmFile("./wasms/standard_payment.wasm"),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(10),
	)

	handlerMsgExecute(input.ctx, input.elk, counterCallMSG)

	nextBlockABCI2 := abci.RequestBeginBlock{
		Hash:   blockState2,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: input.blockStateHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI2, input.elk)

	arrPath := strings.Split(path, "/")

	stateHash1 := input.elk.GetEEState(input.ctx, blockState1)
	res1, _ := grpc.Query(input.elk.client, stateHash1, "address", input.genesisAddress, arrPath, input.elk.protocolVersion)
	assert.Equal(t, int32(0), res1.GetIntValue())

	stateHash2 := input.elk.GetEEState(input.ctx, blockState2)
	res2, _ := grpc.Query(input.elk.client, stateHash2, "address", input.genesisAddress, arrPath, input.elk.protocolVersion)
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
