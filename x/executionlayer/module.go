package executionlayer

import (
	"encoding/json"

	"github.com/gorilla/mux"
	abci "github.com/hdac-io/tendermint/abci/types"
	"github.com/spf13/cobra"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/module"

	"github.com/hdac-io/friday/x/executionlayer/client/cli"
	"github.com/hdac-io/friday/x/executionlayer/client/rest"
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
	rest.RegisterRoutes(ctx, rtr, types.ModuleName)
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return nil
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return nil
}

// BuildCreateValidatorMsg - used for gen-tx
func (AppModuleBasic) BuildCreateValidatorMsg(cliCtx context.CLIContext) (sdk.Msg, error) {
	return cli.BuildCreateValidatorMsg(cliCtx)
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
func (am AppModule) NewHandler() sdk.Handler { return NewHandler(am.keeper) }

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
	var genesisState types.GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)

	return InitGenesis(ctx, am.keeper, genesisState)
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesisState := ExportGenesis(ctx, am.keeper)
	return ModuleCdc.MustMarshalJSON(genesisState)
}

// module begin-block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// module end-block
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, req, am.keeper)
}
