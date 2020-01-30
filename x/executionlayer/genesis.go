package executionlayer

import (
	"reflect"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// InitGenesis sets an executionlayer configuration for genesis.
func InitGenesis(
	ctx sdk.Context, keeper ExecutionLayerKeeper, data types.GenesisState) {
	genesisConfig, err := types.ToChainSpecGenesisConfig(data)
	if err != nil {
		panic(err)
	}

	response, err := keeper.client.RunGenesis(ctx.Context(), genesisConfig)
	if err != nil {
		panic(err)
	}

	if reflect.TypeOf(response.GetResult()) != reflect.TypeOf(&ipc.GenesisResponse_Success{}) {
		panic(response.GetResult())
	}

	if data.Accounts != nil {
		keeper.SetGenesisAccounts(ctx, data.Accounts)
	}
	keeper.SetChainName(ctx, data.ChainName)

	keeper.SetGenesisConf(ctx, data.GenesisConf)

	keeper.SetEEState(ctx, []byte(types.GenesisBlockHashKey), response.GetSuccess().GetPoststateHash())
}

// ExportGenesis : exports an executionlayer configuration for genesis
func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) types.GenesisState {
	return types.NewGenesisState(
		keeper.GetGenesisConf(ctx), keeper.GetGenesisAccounts(ctx), keeper.GetChainName(ctx))
}
