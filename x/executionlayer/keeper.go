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
	"github.com/hdac-io/friday/x/executionlayer/types"
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
// Used when genesis load
func (k ExecutionLayerKeeper) InitialUnitHashMap(ctx sdk.Context, blockHash []byte) {
	emptyHash := util.DecodeHexString(util.StrEmptyStateHash)
	k.SetEEState(ctx, blockHash, emptyHash)
}

// SetUnitHashMap map unitHash to blockHash
func (k ExecutionLayerKeeper) SetUnitHashMap(ctx sdk.Context, blockHash []byte, unitHash UnitHashMap) bool {
	if bytes.Equal(blockHash, []byte{}) {
		blockHash = []byte("genesis")
	}
	if bytes.Equal(unitHash.EEState, []byte{}) || len(unitHash.EEState) != 32 {
		return false
	}

	unitBytes, err := k.cdc.MarshalBinaryBare(unitHash)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set(blockHash, unitBytes)

	return true
}

// GetUnitHashMap returns a UnitHashMap for blockHash
func (k ExecutionLayerKeeper) GetUnitHashMap(ctx sdk.Context, blockHash []byte) UnitHashMap {
	if bytes.Equal(blockHash, []byte{}) {
		blockHash = []byte("genesis")
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	unitBytes := store.Get(blockHash)
	var unit UnitHashMap
	k.cdc.UnmarshalBinaryBare(unitBytes, &unit)
	return unit
}

// SetEEState map eeState to blockHash
func (k ExecutionLayerKeeper) SetEEState(ctx sdk.Context, blockHash []byte, eeState []byte) bool {
	if bytes.Equal(blockHash, []byte{}) {
		blockHash = []byte("genesis")
	}
	if bytes.Equal(eeState, []byte{}) || len(eeState) != 32 {
		return false
	}

	unit := UnitHashMap{
		EEState: eeState,
	}

	unitBytes, err := k.cdc.MarshalBinaryBare(unit)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set(blockHash, unitBytes)

	return true
}

// GetEEState returns a eeState for blockHash
func (k ExecutionLayerKeeper) GetEEState(ctx sdk.Context, blockHash []byte) []byte {
	if bytes.Equal(blockHash, []byte{}) {
		blockHash = []byte("genesis")
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	unitBytes := store.Get(blockHash)
	var unit UnitHashMap
	k.cdc.UnmarshalBinaryBare(unitBytes, &unit)
	return unit.EEState
}

// GetQueryResult queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryResult(ctx sdk.Context,
	stateHash []byte, keyType string, keyData string, path string) (state.Value, error) {
	arrPath := strings.Split(path, "/")

	var changedkeydata []byte
	if keyType == "address" {
		bech32addr, err := sdk.AccAddressFromBech32(keyData)
		if err != nil {
			return state.Value{}, err
		}
		changedkeydata = types.ToPublicKey(bech32addr)
		fmt.Println(changedkeydata)
	} else {
		changedkeydata = util.DecodeHexString(keyData)
	}

	res, err := grpc.Query(k.client, stateHash, keyType, changedkeydata, arrPath, k.protocolVersion)
	if err != "" {
		return state.Value{}, fmt.Errorf(err)
	}

	return *res, nil
}

// GetQueryResultSimple queries without state hash.
// State hash comes from Tendermint block state - EE state mapping DB
func (k ExecutionLayerKeeper) GetQueryResultSimple(ctx sdk.Context,
	keyType string, keyData string, path string) (state.Value, error) {
	unitHash := k.GetUnitHashMap(ctx, k.GetCurrentBlockHash(ctx))
	arrPath := strings.Split(path, "/")

	var changedkeydata []byte
	if keyType == "address" {
		bech32addr, err := sdk.AccAddressFromBech32(keyData)
		if err != nil {
			return state.Value{}, err
		}
		changedkeydata = types.ToPublicKey(bech32addr)
	} else {
		changedkeydata = util.DecodeHexString(keyData)
	}

	res, err := grpc.Query(k.client, unitHash.EEState, keyType, changedkeydata, arrPath, k.protocolVersion)
	if err != "" {
		return state.Value{}, fmt.Errorf(err)
	}

	return *res, nil
}

// GetQueryBalanceResult queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResult(ctx sdk.Context, stateHash []byte, address types.PublicKey) (string, error) {
	res, err := grpc.QueryBalance(k.client, stateHash, address, k.protocolVersion)
	if err != "" {
		return "", fmt.Errorf(err)
	}

	return res, nil
}

// GetQueryBalanceResultSimple queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResultSimple(ctx sdk.Context, address types.PublicKey) (string, error) {
	unitHash := k.GetUnitHashMap(ctx, k.GetCurrentBlockHash(ctx))
	res, err := grpc.QueryBalance(k.client, unitHash.EEState, address, k.protocolVersion)
	if err != "" {
		return "", fmt.Errorf(err)
	}

	return res, nil
}

// GetGenesisConf retrieves GenesisConf from sdk store
func (k ExecutionLayerKeeper) GetGenesisConf(ctx sdk.Context) types.GenesisConf {
	store := ctx.KVStore(k.HashMapStoreKey)
	genesisConfBytes := store.Get([]byte("genesisconf"))

	var genesisConf types.GenesisConf
	k.cdc.UnmarshalBinaryBare(genesisConfBytes, &genesisConf)
	return genesisConf
}

// SetGenesisConf saves GenesisConf in sdk store
func (k ExecutionLayerKeeper) SetGenesisConf(ctx sdk.Context, genesisConf types.GenesisConf) {
	store := ctx.KVStore(k.HashMapStoreKey)
	genesisConfBytes := k.cdc.MustMarshalBinaryBare(genesisConf)
	store.Set([]byte("genesisconf"), genesisConfBytes)
}

// GetCurrentBlockHash returns current block hash
func (k ExecutionLayerKeeper) GetCurrentBlockHash(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.HashMapStoreKey)
	blockHash := store.Get([]byte("currentblockhash"))

	return blockHash
}

// SetCurrentBlockHash saves current block hash
func (k ExecutionLayerKeeper) SetCurrentBlockHash(ctx sdk.Context, blockHash []byte) {
	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set([]byte("currentblockhash"), blockHash)
}
