package executionlayer

import (
	"strconv"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
)

type ExecutionLayerKeeper struct {
	storeKey        sdk.StoreKey
	client          ipc.ExecutionEngineServiceClient
	protocolVersion *state.ProtocolVersion
	cdc             *codec.Codec
}

func NewExecutionLayerKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, path string, protocolVersion string) ExecutionLayerKeeper {
	pv := strings.Split(protocolVersion, ".")
	pvInts := make([]int, 3)
	pvInts[0], _ = strconv.Atoi(pv[0])
	pvInts[1], _ = strconv.Atoi(pv[1])
	pvInts[2], _ = strconv.Atoi(pv[2])
	return ExecutionLayerKeeper{
		storeKey:        storeKey,
		client:          grpc.Connect(path),
		protocolVersion: &state.ProtocolVersion{Major: uint32(pvInts[0]), Minor: uint32(pvInts[1]), Patch: uint32(pvInts[2])},
		cdc:             cdc,
	}
}

func (k Keeper) SetUnitHashMap(ctx sdk.Context, blockState []byte, eeState []byte) {
	if Equal(blockState, []byte{}) {
		return
	}
	if Equal(eeState, []byte{}) || len(eeState) != 32 {
		return
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(blockState, eeState)
}

func (k Keeper) GetEEState(ctx sdk.Context, blockState []byte) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(blockState)
}
