package executionlayer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"

	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
)

type ExecutionLayerKeeper struct {
	HashMapStoreKey sdk.StoreKey
	client          ipc.ExecutionEngineServiceClient
	protocolVersion *state.ProtocolVersion
	cdc             *codec.Codec
}

func NewExecutionLayerKeeper(
	cdc *codec.Codec, hashMapStoreKey sdk.StoreKey, path string, protocolVersion string) ExecutionLayerKeeper {
	pv := strings.Split(protocolVersion, ".")
	pvInts := make([]int, 3)
	pvInts[0], _ = strconv.Atoi(pv[0])
	pvInts[1], _ = strconv.Atoi(pv[1])
	pvInts[2], _ = strconv.Atoi(pv[2])
	return ExecutionLayerKeeper{
		HashMapStoreKey: hashMapStoreKey,
		client:          grpc.Connect(path),
		protocolVersion: &state.ProtocolVersion{Major: uint32(pvInts[0]), Minor: uint32(pvInts[1]), Patch: uint32(pvInts[2])},
		cdc:             cdc,
	}
}

// InitialUnitHashMap initial UnitMapHash using empty hash value
func (k ExecutionLayerKeeper) InitialUnitHashMap(ctx sdk.Context, blockState []byte) {
	emptyHash := util.DecodeHexString(util.StrEmptyStateHash)
	k.SetUnitHashMap(ctx, blockState, emptyHash)
}

// SetUnitHashMap map blockstate eeState to blockState
func (k ExecutionLayerKeeper) SetUnitHashMap(ctx sdk.Context, blockState []byte, eeState []byte) bool {
	if bytes.Equal(blockState, []byte{}) {
		return false
	}
	if bytes.Equal(eeState, []byte{}) || len(eeState) != 32 {
		return false
	}

	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set(blockState, eeState)

	return true
}

// GetEEState returns a eeState for blockState
func (k ExecutionLayerKeeper) GetEEState(ctx sdk.Context, blockState []byte) []byte {
	store := ctx.KVStore(k.HashMapStoreKey)
	eeState := store.Get(blockState)
	return eeState
}

// GetQueryResult queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryResult(ctx sdk.Context,
	stateHash []byte, keyType string, keyData string, path string) (state.Value, error) {
	arrPath := strings.Split(path, "/")
	res, err := grpc.Query(k.client, stateHash, keyType, keyData, arrPath, k.protocolVersion)
	if err != "" {
		return state.Value{}, fmt.Errorf(err)
	}

	return *res, nil
}

// GetQueryResultSimple queries without state hash.
// State hash comes from Tendermint block state - EE state mapping DB
func (k ExecutionLayerKeeper) GetQueryResultSimple(ctx sdk.Context,
	keyType string, keyData string, path string) (state.Value, error) {
	stateHash := k.GetEEState(ctx, ctx.BlockHeader().LastBlockId.Hash)
	arrPath := strings.Split(path, "/")
	res, err := grpc.Query(k.client, stateHash, keyType, keyData, arrPath, k.protocolVersion)
	if err != "" {
		return state.Value{}, fmt.Errorf(err)
	}

	return *res, nil
}

// GetQueryBalanceResult queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResult(ctx sdk.Context, stateHash []byte, address string) (string, error) {
	res, err := grpc.QueryBlanace(k.client, stateHash, address, k.protocolVersion)
	if err != "" {
		return "", fmt.Errorf(err)
	}

	return res, nil
}

// GetQueryBalanceResultSimple queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResultSimple(ctx sdk.Context, address string) (string, error) {
	stateHash := k.GetEEState(ctx, ctx.BlockHeader().LastBlockId.Hash)
	res, err := grpc.QueryBlanace(k.client, stateHash, address, k.protocolVersion)
	if err != "" {
		return "", fmt.Errorf(err)
	}

	return res, nil
}
