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

	response, err := keeper.client.RunGenesis(ctx.Context(), genesisConfig)
	if err != nil {
		panic(err)
	}

	if reflect.TypeOf(response.GetResult()) != reflect.TypeOf(&ipc.GenesisResponse_Success{}) {
		panic(response.GetResult())
	}

	stateHash, bonds, errStr := grpc.Commit(keeper.client, util.DecodeHexString(util.StrEmptyStateHash), response.GetSuccess().GetEffect().GetTransformMap(), genesisConfig.GetProtocolVersion())
	if errStr != "" {
		panic(errStr)
	}

	if data.Accounts != nil {
		keeper.SetGenesisAccounts(ctx, data.Accounts)
	}
	keeper.SetChainName(ctx, data.ChainName)

	keeper.SetGenesisConf(ctx, data.GenesisConf)

	keeper.SetEEState(ctx, []byte(types.GenesisBlockHashKey), stateHash)

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.Hash = []byte(types.GenesisBlockHashKey)
	candidateBlock.State = stateHash
	candidateBlock.Bonds = bonds

	systemAccount := make([]byte, 32)
	res, errStr := grpc.Query(keeper.client, stateHash, "address", systemAccount, []string{}, genesisConfig.GetProtocolVersion())
	if errStr != "" {
		panic(errStr)
	}

	storedValue := util.UnmarshalStoreValue(res)
	proxyContractHash := []byte{}
	for _, namedKey := range storedValue.GetAccount().GetNamedKeys() {
		if namedKey.GetName() == types.ProxyContractName {
			proxyContractHash = namedKey.GetKey().GetHash().GetHash()
			break
		}
	}

	if len(proxyContractHash) != 32 {
		panic(fmt.Sprintf("%s must exist. Check systemcontract.", types.ProxyContractName))
	}

	keeper.SetProxyContractHash(ctx, proxyContractHash)
}

// ExportGenesis : exports an executionlayer configuration for genesis
func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) types.GenesisState {
	return types.NewGenesisState(
		keeper.GetGenesisConf(ctx), keeper.GetGenesisAccounts(ctx), keeper.GetChainName(ctx))
}
