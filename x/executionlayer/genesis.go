package executionlayer

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

// InitGenesis sets an executionlayer configuration for genesis.
func InitGenesis(
	ctx sdk.Context, keeper ExecutionLayerKeeper, data types.GenesisState) {
	genesisConfig, err := types.ToChainSpecGenesisConfig(data)
	if err != nil {
		panic(err)
	}
	genesisConfig.Timestamp = uint64(ctx.BlockTime().Unix())

	response, err := keeper.client.RunGenesis(ctx.Context(), genesisConfig)
	if err != nil {
		panic(err)
	}

	if reflect.TypeOf(response.GetResult()) != reflect.TypeOf(&ipc.GenesisResponse_Success{}) {
		panic(response.GetResult())
	}

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.State = response.GetSuccess().PoststateHash

	keeper.SetChainName(ctx, data.ChainName)
	keeper.SetGenesisConf(ctx, data.GenesisConf)
	keeper.SetUnitHashMap(ctx, types.NewUnitHashMap(ctx.CandidateBlock().State))

	// Query to current validator information.
	posInfos, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, types.PosContractName)
	if err != nil {
		panic(err)
	}

	bonds := []*ipc.Bond{}
	validatorStakeInfos := posInfos.Contract.NamedKeys.GetAllValidators()

	for _, validator := range data.Validators {
		bond := &ipc.Bond{
			ValidatorPublicKey: validator.OperatorAddress,
			Stake:              &state.BigInt{Value: validatorStakeInfos[hex.EncodeToString(validator.OperatorAddress)], BitWidth: 512},
		}
		bonds = append(bonds, bond)

		validator.Stake = ""
		keeper.SetValidator(ctx, validator.OperatorAddress, validator)
	}
	candidateBlock.Bonds = bonds

	// initial proxy contract
	storedValueSystemAccount, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, "")
	if err != nil {
		panic(err)
	}

	proxyContractHash := []byte{}
	for _, namedKey := range storedValueSystemAccount.Account.NamedKeys {
		if namedKey.Name == types.ProxyContractName {
			proxyContractHash = namedKey.Key.Hash
			break
		}
	}

	if len(proxyContractHash) != 32 {
		panic(fmt.Sprintf("%s must exist. Check systemcontract.", types.ProxyContractName))
	}

	keeper.SetProxyContractHash(ctx, proxyContractHash)
	keeper.SetUnitHashMap(ctx, types.NewUnitHashMap(ctx.CandidateBlock().State))
}

func ExportGenesis(ctx sdk.Context, keeper ExecutionLayerKeeper) types.GenesisState {
	validators := keeper.GetAllValidators(ctx)
	existAccounts := keeper.AccountKeeper.GetAllAccounts(ctx)

	stateHash := keeper.GetUnitHashMap(ctx, ctx.BlockHeight()).EEState
	protocolVersion := keeper.MustGetProtocolVersion(ctx)

	stakeAmounts := map[string]string{}
	for _, validator := range validators {
		stakeAmounts[validator.OperatorAddress.String()] = validator.Stake
	}

	accounts := []types.Account{}
	for _, existAccount := range existAccounts {
		if strings.Contains(existAccount.String(), "name") {
			continue
		}

		balance, err := grpc.QueryBalance(keeper.client, stateHash, existAccount.GetAddress(), &protocolVersion)
		if err != "" {
			panic(err)
		}

		account := types.Account{
			Address:             existAccount.GetAddress(),
			InitialBalance:      balance,
			InitialBondedAmount: stakeAmounts[existAccount.GetAddress().String()],
		}
		accounts = append(accounts, account)
	}

	stateInfos := []string{}
	if len(stateHash) != 0 {
		systeAccountInfo, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, types.PosContractName)
		if err != nil {
			panic(err)
		}

		for _, namedKey := range systeAccountInfo.Contract.NamedKeys {
			switch namedKey.Name[:2] {
			case storedvalue.DELEGATE_PREFIX + "_", storedvalue.VOTE_PREFIX + "_", storedvalue.REWARD_PREFIX + "_", storedvalue.COMMISSION_PREFIX + "_":
				stateInfos = append(stateInfos, namedKey.Name)
			default:
				continue
			}
		}
	}

	return types.NewGenesisState(
		keeper.GetGenesisConf(ctx), accounts, keeper.GetChainName(ctx), validators, stateInfos)
}
