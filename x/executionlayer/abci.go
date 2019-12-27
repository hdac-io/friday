package executionlayer

import (
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"
)

func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, elk ExecutionLayerKeeper) {
	preHash := req.Header.LastBlockId.Hash
	unitHash := elk.GetUnitHashMap(ctx, preHash)

	elk.SetCandidateBlockHash(ctx, req.Hash)
	elk.SetUnitHashMap(ctx, req.Hash, unitHash)
}

// func EndBlocker(ctx sdk.Context, elk ExecutionLayerKeeper) []abci.ValidatorUpdate {
// 	bonds := elk.GetCandidateBlockBond()

// 	var validatorUpdate = []abci.ValidatorUpdate{}

// 	for _, value := range bonds {
// 		val := abci.ValidatorUpdate{
// 			PubKey: abci.PubKey{Data: value.GetValidatorPublicKey()},
// 			Power:  uint64(value.GetStake().GetValue()),
// 		}
// 	}
// }
