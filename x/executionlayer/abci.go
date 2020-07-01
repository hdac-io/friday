package executionlayer

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	abci "github.com/hdac-io/tendermint/abci/types"
	tmtypes "github.com/hdac-io/tendermint/types"
)

func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, elk ExecutionLayerKeeper) {
	unitHash := elk.GetUnitHashMap(ctx, req.GetHeader().Height-1)

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.Hash = req.GetHash()
	candidateBlock.State = unitHash.EEState
	candidateBlock.Deploys = []*ipc.DeployItem{}
	candidateBlock.Effects = []*transforms.TransformEntry{}
	protocolVersion := elk.GetProtocolVersion(ctx)
	candidateBlock.ProtocolVersion = &protocolVersion
}

func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k ExecutionLayerKeeper) []abci.ValidatorUpdate {

	// Commit
	stateHash, _, errGrpc := grpc.Commit(k.client, ctx.CandidateBlock().State, ctx.CandidateBlock().Effects, ctx.CandidateBlock().ProtocolVersion)
	if errGrpc != "" {
		panic(errGrpc)
	}

	var validatorUpdates []abci.ValidatorUpdate

	// step
	stepRequest := &ipc.StepRequest{
		ParentStateHash: stateHash,
		BlockTime:       uint64(ctx.BlockTime().Unix()),
		BlockHeight:     ctx.UBlockHeight(),
		ProtocolVersion: ctx.CandidateBlock().ProtocolVersion,
	}
	res, err := k.client.Step(ctx.Context(), stepRequest)
	if err != nil {
		panic(err)
	}
	switch res.GetResult().(type) {
	case *ipc.StepResponse_Success:
		ctx.CandidateBlock().State = res.GetSuccess().GetPostStateHash()
	case *ipc.StepResponse_MissingParent:
		panic(fmt.Sprintf("Missing parent : %s", hex.EncodeToString(res.GetMissingParent().GetHash())))
	case *ipc.StepResponse_Error:
		panic(res.GetError().GetMessage())
	default:
		panic(fmt.Sprintf("Unknown result : %s", res.String()))
	}

	// Query to current validator information.
	resPosInfoBytes, err := getQueryResult(ctx, k, types.ADDRESS, types.SYSTEM, types.PosContractName)
	var posInfos storedvalue.StoredValue
	posInfos, err, _ = posInfos.FromBytes(resPosInfoBytes)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	nextStakeInfos := posInfos.Contract.NamedKeys.GetAllValidators()

	// calculate and set voting power
	validators := k.GetAllValidators(ctx)

	if len(nextStakeInfos) > 0 {
		for _, validator := range validators {
			var power string
			stake, found := nextStakeInfos[hex.EncodeToString(validator.OperatorAddress)]
			if found {
				if validator.Stake == stake {
					continue
				}
				power = stake
				validator.Stake = stake
			} else {
				if validator.Stake != "" {
					power = "0"
					validator.Stake = ""
				} else {
					continue
				}
			}

			if len(power) <= types.DECIMAL_POINT_POS {
				power = "0"
			} else {
				power = power[:len(power)-types.DECIMAL_POINT_POS]
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

	unitHash := NewUnitHashMap(ctx.CandidateBlock().State)

	k.SetUnitHashMap(ctx, unitHash)

	return validatorUpdates
}
