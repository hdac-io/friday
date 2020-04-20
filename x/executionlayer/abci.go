package executionlayer

import (
	"strconv"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
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
}

func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k ExecutionLayerKeeper) []abci.ValidatorUpdate {
	var validatorUpdates []abci.ValidatorUpdate

	// step
	proxyContractHash := k.GetProxyContractHash(ctx)
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_StringValue{
					StringValue: types.StepMethodName}}}}

	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		types.SYSTEM,
		types.SYSTEM_ACCOUNT,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		types.BASIC_FEE,
		types.BASIC_GAS,
	)
	result, log := execute(ctx, k, msgExecute)
	if !result {
		getResult(result, log)
	}

	// calculate and set voting power
	validators := k.GetAllValidators(ctx)

	resultbonds := ctx.CandidateBlock().Bonds
	if len(resultbonds) > 0 {
		resultBondsMap := make(map[string]*ipc.Bond)
		for _, bond := range resultbonds {
			resultBondsMap[string(bond.GetValidatorPublicKey())] = bond
		}

		for _, validator := range validators {
			var power string
			resultBond, found := resultBondsMap[string(validator.OperatorAddress.ToEEAddress())]
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
