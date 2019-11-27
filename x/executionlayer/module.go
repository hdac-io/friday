package executionlayer

import (
	"encoding/json"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/module"

	"github.com/hdac-io/friday/x/executionlayer/client/cli"
	"github.com/hdac-io/friday/x/executionlayer/config"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module basics object
type AppModuleBasic struct{}

// module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// module validate genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return types.ValidateGenesis(data)
}

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	// rest.RegisterRoutes(ctx, rtr, types.StoreKey)
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetExecutionLayerTxCmd(cdc)
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetExecutionLayerQueryCmd(cdc)
}

//___________________________
// app module object
type AppModule struct {
	AppModuleBasic
	keeper ExecutionLayerKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(executionLayerKeeper ExecutionLayerKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         executionLayerKeeper,
	}
}

// module name
func (AppModule) Name() string {
	return ModuleName
}

// register invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route works to route msg to this module (revised)
func (AppModule) Route() string { return RouterKey }

// module handler
func (AppModule) NewHandler() sdk.Handler { return nil }

// QuerierRoute works to route query to this module (revised)
func (AppModule) QuerierRoute() string {
	return ModuleName
}

// NewQuerierHandler constructs the query router
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	// var genesisState types.GenesisState
	// ModuleCdc.MustUnmarshalJSON(data, &genesisState)

	// TODO : make GenesisState injectable
	accounts := make([]types.Account, 1)
	accounts[0] = types.Account{
		PublicKey:           "s8qP7TauBe0WoHUDEKyFR99XM6q7aGzacLa6M6vHtO0=",
		InitialBalance:      "50000000000",
		InitialBondedAmount: "1000000",
	}
	genesisState := types.NewGenesisState(accounts)

	// TODO : Remove ReadGenesisConfig Call from here
	genesisConfig, err := config.ReadGenesisConfig(os.ExpandEnv("$HOME/.fryd/chainspec/genesis/manifest.toml"))
	if err != nil {
		panic(err)
	}

	InitGenesis(ctx, am.keeper, genesisState, *genesisConfig)
	return []abci.ValidatorUpdate{}
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return json.RawMessage{}
}

// module begin-block
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// module end-block
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
