package executionlayer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"path"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	abci "github.com/hdac-io/tendermint/abci/types"
	"github.com/stretchr/testify/assert"
)

//-------------------------------------------

func TestQueryKeyToBytes(t *testing.T) {
	_, err := toBytes("address", "friday1dl2cjlfpmc9hcyd4rxts047tze87s0gxmzqx70")
	assert.Nil(t, err)
	_, err = toBytes("address", "invalid address")
	assert.NotNil(t, err)

	expected := []byte("test-data")

	got, err := toBytes("uref", hex.EncodeToString(expected))
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
	_, err = toBytes("hash", hex.EncodeToString(expected))
	assert.Nil(t, err)
	assert.Equal(t, expected, got)
	_, err = toBytes("local", hex.EncodeToString(expected))
	assert.Nil(t, err)
	assert.Equal(t, expected, got)

	_, err = toBytes("invalid key type", "")
	assert.True(t, strings.Contains(err.Error(), "Unknown QueryKey type:"))
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

func TestMustGetProtocolVersion(t *testing.T) {
	expected, err := types.ToProtocolVersion(types.DefaultGenesisState().GenesisConf.Genesis.ProtocolVersion)
	assert.Nil(t, err)

	input := setupTestInput()
	got := input.elk.MustGetProtocolVersion(input.ctx)
	assert.Equal(t, *expected, got)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustGetProtocolVersion below should panic!")
		}
	}()
	input.elk.MustGetProtocolVersion(sdk.Context{})
}

func TestCreateBlock(t *testing.T) {
	input := setupTestInput()
	parentHash := genesis(input.elk)
	blockHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	input.elk.SetEEState(input.ctx, blockHash, parentHash)
	queryPath := "counter/count"

	blockHash1 := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	blockHash2 := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	counterDefineMSG := NewMsgExecute(
		blockHash,
		GenesisAccountAddress,
		GenesisAccountAddress,
		util.LoadWasmFile(path.Join(contractPath, counterDefineWasm)),
		[]byte{},
		util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm)),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(10),
	)

	handlerMsgExecute(input.ctx, input.elk, counterDefineMSG)

	nextBlockABCI1 := abci.RequestBeginBlock{
		Hash:   blockHash1,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: blockHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI1, input.elk)

	counterCallMSG := NewMsgExecute(
		blockHash,
		GenesisAccountAddress,
		GenesisAccountAddress,
		util.LoadWasmFile(path.Join(contractPath, counterCallWasm)),
		[]byte{},
		util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm)),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(10),
	)

	handlerMsgExecute(input.ctx, input.elk, counterCallMSG)

	nextBlockABCI2 := abci.RequestBeginBlock{
		Hash:   blockHash2,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: blockHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI2, input.elk)

	arrPath := strings.Split(queryPath, "/")

	unitHash1 := input.elk.GetUnitHashMap(input.ctx, blockHash1)
	pv := input.elk.MustGetProtocolVersion(input.ctx)
	res1, _ := grpc.Query(input.elk.client, unitHash1.EEState, "address", types.ToPublicKey(GenesisAccountAddress), arrPath, &pv)
	assert.Equal(t, int32(0), res1.GetIntValue())

	unitHash2 := input.elk.GetUnitHashMap(input.ctx, blockHash2)
	res2, _ := grpc.Query(input.elk.client, unitHash2.EEState, "address", types.ToPublicKey(GenesisAccountAddress), arrPath, &pv)
	assert.Equal(t, int32(1), res2.GetIntValue())
}

func TestTransfer(t *testing.T) {
	input := setupTestInput()
	parentHash := genesis(input.elk)
	blockHash := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	input.elk.SetEEState(input.ctx, blockHash, parentHash)

	blockHash1 := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	nextBlockABCI1 := abci.RequestBeginBlock{
		Hash:   blockHash1,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: blockHash}},
	}

	BeginBlocker(input.ctx, nextBlockABCI1, input.elk)

	transferMSG := NewMsgTransfer(
		GenesisAccountAddress,
		GenesisAccountAddress,
		RecipientAccountAddress,
		util.LoadWasmFile(path.Join(contractPath, transferWasm)),
		util.MakeArgsTransferToAccount(types.ToPublicKey(RecipientAccountAddress), 100000000),
		util.LoadWasmFile(path.Join(contractPath, standardPaymentWasm)),
		util.MakeArgsStandardPayment(new(big.Int).SetUint64(200000000)),
		uint64(200000000),
	)

	handlerMsgTransfer(input.ctx, input.elk, transferMSG)

	blockHash2 := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	nextBlockABCI2 := abci.RequestBeginBlock{
		Hash:   blockHash2,
		Header: abci.Header{LastBlockId: abci.BlockID{Hash: blockHash1}},
	}

	BeginBlocker(input.ctx, nextBlockABCI2, input.elk)

	res, err := input.elk.GetQueryBalanceResultSimple(input.ctx, types.ToPublicKey(RecipientAccountAddress))
	queriedRes, _ := strconv.Atoi(res)

	assert.Equal(t, queriedRes, 100000000)
	assert.Equal(t, err, nil)

	res2, err := input.elk.GetQueryBalanceResultSimple(input.ctx, types.ToPublicKey(GenesisAccountAddress))
	queriedRes2, _ := strconv.Atoi(res2)
	fmt.Println(queriedRes)
	fmt.Println(queriedRes2)
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
	var got types.GenesisState

	// GenesisConf test
	testMock.elk.SetGenesisConf(testMock.ctx, expected.GenesisConf)
	got.GenesisConf = testMock.elk.GetGenesisConf(testMock.ctx)

	if !reflect.DeepEqual(expected.GenesisConf.WasmCosts, got.GenesisConf.WasmCosts) {
		t.Errorf("expected: %v, but got: %v", expected.GenesisConf.WasmCosts, got.GenesisConf.WasmCosts)
	}
	assert.Equal(t, expected.GenesisConf.Genesis.Timestamp, got.GenesisConf.Genesis.Timestamp)
	assert.Equal(t, expected.GenesisConf.Genesis.ProtocolVersion, got.GenesisConf.Genesis.ProtocolVersion)
	assert.Equal(t, expected.GenesisConf.Genesis.MintWasm, got.GenesisConf.Genesis.MintWasm)
	assert.Equal(t, expected.GenesisConf.Genesis.PosWasm, got.GenesisConf.Genesis.PosWasm)

	// GenesisAccounts test
	expected.Accounts = make([]types.Account, 1)
	expected.Accounts[0].PublicKey = types.PublicKey([]byte("test-pub-key"))
	expected.Accounts[0].InitialBalance = "2"
	expected.Accounts[0].InitialBondedAmount = "1"

	testMock.elk.SetGenesisAccounts(testMock.ctx, expected.Accounts)
	gottonAccounts := testMock.elk.GetGenesisAccounts(testMock.ctx)
	if !reflect.DeepEqual(expected.Accounts, gottonAccounts) {
		t.Errorf("expected: %v, but got: %v", expected.Accounts, gottonAccounts)
	}

	// ChainName test
	expected.ChainName = "keeper-test-chain-name"
	testMock.elk.SetChainName(testMock.ctx, expected.ChainName)
	gottonChainName := testMock.elk.GetChainName(testMock.ctx)
	assert.Equal(t, expected.ChainName, gottonChainName)
}
