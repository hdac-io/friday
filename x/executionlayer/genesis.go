package executionlayer

import (
	"reflect"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(
	ctx sdk.Context, keeper ExecutionLayerKeeper, data types.GenesisState,
	genesisConfig ipc.ChainSpec_GenesisConfig) {
	// Genesis Accounts
	if accountsLen := len(data.Accounts); accountsLen != 0 {
		genesisConfig.Accounts = make([]*ipc.ChainSpec_GenesisAccount, accountsLen)
		for i, v := range data.Accounts {
			genAccount, err := types.ToGenesisAccount(v)
			if err != nil {
				panic(err)
			}
			genesisConfig.Accounts[i] = genAccount
		}
	}

	response, err := grpc.RunGenesis(keeper.client, &genesisConfig)
	if err != nil {
		panic(err)
	}

	if reflect.TypeOf(response.GetResult()) != reflect.TypeOf(&ipc.GenesisResponse_Success{}) {
		panic(response.GetResult())
	}

	// keeper.InitialUnitHashMap(ctx)
	//
	// keeper.SetUnitHashMap(ctx, ctx.BlockHeader().LastBlockId.GetHash(), reponse.poststatehash)
}

// ExportGenesis :
func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) error {
	return nil
}
