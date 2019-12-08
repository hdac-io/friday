package executionlayer

import (
	"reflect"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// InitGenesis sets an executionlayer configuration for genesis.
func InitGenesis(
	ctx sdk.Context, keeper ExecutionLayerKeeper, data types.GenesisState) {
	keeper.InitialUnitHashMap(ctx, ctx.BlockHeader().LastBlockId.Hash)
	genesisConfig, err := types.ToChainSpecGenesisConfig(data.GenesisConf)
	if err != nil {
		panic(err)
	}
	response, err := grpc.RunGenesis(keeper.client, genesisConfig)
	if err != nil {
		panic(err)
	}

	if reflect.TypeOf(response.GetResult()) != reflect.TypeOf(&ipc.GenesisResponse_Success{}) {
		panic(response.GetResult())
	}

	keeper.SetGenesisConf(ctx, data.GenesisConf)
	keeper.SetEEState(ctx, ctx.BlockHeader().LastBlockId.Hash, response.GetSuccess().GetPoststateHash())
}

// ExportGenesis : exports an executionlayer configuration for genesis
func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) types.GenesisState {
	return types.NewGenesisState(keeper.GetGenesisConf(ctx))
}
