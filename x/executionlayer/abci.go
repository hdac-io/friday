package executionlayer

import (
	"strconv"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"
	tmtypes "github.com/hdac-io/tendermint/types"
)

func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, elk ExecutionLayerKeeper) {
	preHash := req.Header.LastBlockId.Hash
	unitHash := elk.GetUnitHashMap(ctx, preHash)

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.Hash = req.GetHash()
	candidateBlock.State = unitHash.EEState
}

func EndBloker(ctx sdk.Context, k ExecutionLayerKeeper) []abci.ValidatorUpdate {
	var validatorUpdates []abci.ValidatorUpdate

	validators := k.GetAllValidators(ctx)

	resultbonds := ctx.CandidateBlock().Bonds
	resultBondsMap := make(map[string]*ipc.Bond)
	for _, bond := range resultbonds {
		resultBondsMap[string(bond.GetValidatorPublicKey())] = bond
	}

	if len(resultbonds) > 0 {
		for _, validator := range validators {
			var power string
			resultBond, found := resultBondsMap[string(validator.OperatorAddress.Bytes())]
			if found {
				if validator.Stake == resultBond.GetStake().GetValue() {
					continue
				}
				power = resultBond.GetStake().GetValue()
				validator.Stake = resultBond.GetStake().GetValue()
			} else {
				if validator.Stake != "" {
					power = "0"
					validator.Stake = ""
				} else {
					continue
				}
			}
			coin, err := strconv.ParseInt(power, 10, 64)
			if err != nil {
				continue
			}
			validatorUpdate := abci.ValidatorUpdate{
				PubKey: tmtypes.TM2PB.PubKey(validator.ConsPubKey),
				Power:  coin,
			}
			validatorUpdates = append(validatorUpdates, validatorUpdate)
			k.SetValidator(ctx, validator.OperatorAddress, validator)
		}
	}

	k.SetEEState(ctx, ctx.CandidateBlock().Hash, ctx.CandidateBlock().State)

	return validatorUpdates
}
