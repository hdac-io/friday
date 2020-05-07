package executionlayer

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
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

	// add to temp account
	tempAddress := sdk.AccAddress(types.TEMP_ACC_ADDRESS)
	account := &ipc.ChainSpec_GenesisAccount{
		PublicKey:    tempAddress,
		Balance:      &state.BigInt{Value: types.SYSTEM_ACCOUNT_BALANCE, BitWidth: 512},
		BondedAmount: &state.BigInt{Value: types.SYSTEM_ACCOUNT_BONDED_AMOUNT, BitWidth: 512},
	}
	genesisConfig.Accounts = append(genesisConfig.Accounts, account)

	response, err := keeper.client.RunGenesis(ctx.Context(), genesisConfig)
	if err != nil {
		panic(err)
	}

	if reflect.TypeOf(response.GetResult()) != reflect.TypeOf(&ipc.GenesisResponse_Success{}) {
		panic(response.GetResult())
	}

	if data.Accounts != nil {
		keeper.SetGenesisAccounts(ctx, data.Accounts)
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
			Stake:              &state.BigInt{Value: validatorStakeInfos[hex.EncodeToString(validator.OperatorAddress.ToEEAddress())], BitWidth: 512},
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

	// send to system account from temp account
	sessionArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.TransferMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_ListType{ListType: &state.CLType_List{Inner: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U8}}}}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_BytesValue{
						BytesValue: types.SYSTEM_ACCOUNT}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: types.TRANSFER_BALANCE}}}}}}
	sessionArgsStr, err := DeployArgsToJsonString(sessionArgs)
	if err != nil {
		getResult(false, err.Error())
	}

	msgExecute := NewMsgExecute(
		types.SYSTEM,
		tempAddress,
		util.HASH,
		proxyContractHash,
		sessionArgsStr,
		types.BASIC_FEE,
	)
	result, log := execute(ctx, keeper, msgExecute)
	if !result {
		getResult(false, log)
	}

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

		balance, err := grpc.QueryBalance(keeper.client, stateHash, existAccount.GetAddress().ToEEAddress(), &protocolVersion)
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
