package executionlayer

import (
	"fmt"
	"reflect"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
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
	stateHash, bonds, errStr := grpc.Commit(keeper.client, util.DecodeHexString(util.StrEmptyStateHash), response.GetSuccess().GetEffect().GetTransformMap(), genesisConfig.GetProtocolVersion())
	if errStr != "" {
		panic(errStr)
	}

	keeper.SetGenesisConf(ctx, data.GenesisConf)

	candidateBlock := types.CandidateBlock{
		Hash:  []byte(types.GenesisBlockHashKey),
		Bonds: bonds,
	}
	keeper.SetCandidateBlock(ctx, candidateBlock)
	keeper.SetEEState(ctx, []byte(types.GenesisBlockHashKey), stateHash)
}

// ExportGenesis : exports an executionlayer configuration for genesis
func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) types.GenesisState {
	return types.NewGenesisState(
		keeper.GetGenesisConf(ctx), keeper.GetGenesisAccounts(ctx), keeper.GetChainName(ctx))
}
