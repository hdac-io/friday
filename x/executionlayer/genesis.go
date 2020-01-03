package executionlayer

import (
	"fmt"
	"reflect"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
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

	isMintValid, _ := grpc.Validate(
		keeper.client, genesisConfig.GetMintInstaller(), genesisConfig.GetProtocolVersion())
	isPosValid, _ := grpc.Validate(
		keeper.client, genesisConfig.GetPosInstaller(), genesisConfig.GetProtocolVersion())

	if !isMintValid || !isPosValid {
		panic(fmt.Errorf("Bad system contracts. mint: %v, pos: %v", isMintValid, isPosValid))
	}

	response, err := grpc.RunGenesis(keeper.client, genesisConfig)
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
	keeper.SetEEState(ctx, ctx.BlockHeader().LastBlockId.Hash, response.GetSuccess().GetPoststateHash())
}

// ExportGenesis : exports an executionlayer configuration for genesis
func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) types.GenesisState {
	return types.NewGenesisState(
		keeper.GetGenesisConf(ctx), keeper.GetGenesisAccounts(ctx), keeper.GetChainName(ctx))
}
