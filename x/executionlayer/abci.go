package executionlayer

import (
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, elk ExecutionLayerKeeper) {
	preHash := req.Header.LastBlockId.Hash
	eeState := elk.GetEEState(ctx, preHash)

	elk.SetUnitHashMap(ctx, req.Hash, eeState)
}
