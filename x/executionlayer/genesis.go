package executionlayer

import (
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

	stateHash, bonds, errStr := grpc.Commit(keeper.client, util.DecodeHexString(util.StrEmptyStateHash), response.GetSuccess().GetEffect().GetTransformMap(), genesisConfig.GetProtocolVersion())
	if errStr != "" {
		panic(errStr)
	}

	if data.Accounts != nil {
		keeper.SetGenesisAccounts(ctx, data.Accounts)
	}
	keeper.SetChainName(ctx, data.ChainName)
	keeper.SetGenesisConf(ctx, data.GenesisConf)

	for _, validator := range data.Validators {
		validator.Stake = ""
		keeper.SetValidator(ctx, validator.OperatorAddress, validator)
	}

	candidateBlock := ctx.CandidateBlock()
	candidateBlock.State = stateHash
	candidateBlock.Bonds = bonds

	// initial proxy contract
	res, errStr := grpc.Query(keeper.client, stateHash, "address", types.SYSTEM_ACCOUNT, []string{}, genesisConfig.GetProtocolVersion())
	if errStr != "" {
		panic(errStr)
	}

	var storedValue storedvalue.StoredValue
	storedValue, err, _ = storedValue.FromBytes(res)
	if err != nil {
		panic(err)
	}

	proxyContractHash := []byte{}
	for _, namedKey := range storedValue.Account.NamedKeys {
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
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_StringValue{
					StringValue: types.TransferMethodName}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_BytesValue{
					BytesValue: types.SYSTEM_ACCOUNT}}},
		&consensus.Deploy_Arg{
			Value: &consensus.Deploy_Arg_Value{
				Value: &consensus.Deploy_Arg_Value_BigInt{
					BigInt: &state.BigInt{Value: types.TRANSFER_BALANCE, BitWidth: 512}}}},
	}
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
		types.BASIC_GAS,
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

	return types.NewGenesisState(
		keeper.GetGenesisConf(ctx), accounts, keeper.GetChainName(ctx), validators)
}
