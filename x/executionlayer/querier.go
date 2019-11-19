package executionlayer

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/hdac-io/friday/types"
)

// creates a querier for auth REST endpoints
func NewQuerier(keeper ExecutionLayerKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		default:
			return nil, sdk.ErrUnknownRequest("unknown auth query endpoint")
		}
	}
}
