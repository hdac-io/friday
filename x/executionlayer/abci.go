package executionlayer

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"

	"github.com/Workiva/go-datastructures/queue"
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
	protocolVersion := elk.GetProtocolVersion(ctx)
	candidateBlock.ProtocolVersion = &protocolVersion
	candidateBlock.TxsCount = req.Header.GetNumTxs()
	candidateBlock.DeployPQueue = queue.NewPriorityQueue(int(candidateBlock.TxsCount), false)
	candidateBlock.NewAccounts = queue.NewPriorityQueue(int(candidateBlock.TxsCount), false)

	if candidateBlock.TxsCount > 0 {
		candidateBlock.WaitGroup = sync.WaitGroup{}
		candidateBlock.WaitGroup.Add(int(candidateBlock.TxsCount))
		ctx = ctx.WithCandidateBlock(candidateBlock)

		ctx.CandidateBlock().CurrentTxIndex = 0

		var mutex = new(sync.Mutex)
		candidateBlock.AnteCond = sync.NewCond(mutex)
	}
}

func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k ExecutionLayerKeeper) []abci.ValidatorUpdate {
	stateHash := ctx.CandidateBlock().State

	if ctx.CandidateBlock().TxsCount > 0 {
		ctx.CandidateBlock().WaitGroup.Wait()

		deploys := []*ipc.DeployItem{}

		itemDeploysList, err := ctx.CandidateBlock().DeployPQueue.Get(ctx.CandidateBlock().DeployPQueue.Len())

		for _, item := range itemDeploysList {
			itemDeploy := item.(*sdk.ItemDeploy)
			deploys = append(deploys, itemDeploy.Deploy)
		}

		// Execute
		reqExecute := &ipc.ExecuteRequest{
			ParentStateHash: stateHash,
			BlockTime:       uint64(ctx.BlockTime().Unix()),
			Deploys:         deploys,
			ProtocolVersion: ctx.CandidateBlock().ProtocolVersion,
		}

		resExecute, err := k.client.Execute(ctx.Context(), reqExecute)
		if err != nil {
			panic(err)
		}

		effects := []*transforms.TransformEntry{}
		switch resExecute.GetResult().(type) {
		case *ipc.ExecuteResponse_Success:
			for index, res := range resExecute.GetSuccess().GetDeployResults() {
				switch res.GetExecutionResult().GetError().GetValue().(type) {
				case *ipc.DeployError_GasError:
					err = types.ErrGRpcExecuteDeployGasError(types.DefaultCodespace)
				case *ipc.DeployError_ExecError:
					err = types.ErrGRpcExecuteDeployExecError(types.DefaultCodespace, res.GetExecutionResult().GetError().GetExecError().GetMessage())
				}

				effects = append(effects, res.GetExecutionResult().GetEffects().GetTransformMap()...)
				if err != nil {
					itemDeploysList[index].(*sdk.ItemDeploy).LogChannel <- err.Error()
				} else {
					itemDeploysList[index].(*sdk.ItemDeploy).LogChannel <- ""
				}
			}

		case *ipc.ExecuteResponse_MissingParent:
			err = types.ErrGRpcExecuteMissingParent(types.DefaultCodespace, hex.EncodeToString(resExecute.GetMissingParent().GetHash()))
			for _, itemDeploy := range itemDeploysList {
				itemDeploy.(*sdk.ItemDeploy).LogChannel <- err.Error()
			}
		default:
			err = fmt.Errorf("Unknown result : %s", resExecute.String())
			for _, itemDeploy := range itemDeploysList {
				itemDeploy.(*sdk.ItemDeploy).LogChannel <- err.Error()
			}
		}

		// Commit
		errGrpc := ""
		stateHash, _, errGrpc = grpc.Commit(k.client, ctx.CandidateBlock().State, effects, ctx.CandidateBlock().ProtocolVersion)
		if errGrpc != "" {
			panic(errGrpc)
		}
	}

	// Add to new accounts
	newAccounts, err := ctx.CandidateBlock().NewAccounts.Get(ctx.CandidateBlock().NewAccounts.Len())

	for _, addr := range newAccounts {
		acc, _ := sdk.AccAddressFromBech32(string(*addr.(*sdk.StrAccAddress)))
		toAddressAccountObject := k.AccountKeeper.NewAccountWithAddress(ctx, acc)
		k.AccountKeeper.SetAccount(ctx, toAddressAccountObject)
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
