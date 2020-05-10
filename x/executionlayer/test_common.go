package executionlayer

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/store"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	authtypes "github.com/hdac-io/friday/x/auth/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/hdac-io/friday/x/nickname"
	"github.com/hdac-io/friday/x/params/subspace"
	abci "github.com/hdac-io/tendermint/abci/types"
	"github.com/hdac-io/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const (
	chainID = "test-chain-id"
)

var (
	ContractAddress            = "friday15evpva2u57vv6l5czehyk1111111111111"
	GenesisAccountAddress, _   = sdk.AccAddressFromBech32("friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz")
	RecipientAccountAddress, _ = sdk.AccAddressFromBech32("friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv")

	contractPath        = os.ExpandEnv("$HOME/.nodef/contracts")
	mintInstallWasm     = "mint_install.wasm"
	posInstallWasm      = "pos_install.wasm"
	standardPaymentWasm = "standard_payment.wasm"
	counterDefineWasm   = "counter_define.wasm"
	counterCallWasm     = "counter_call.wasm"
	transferWasm        = "transfer_to_account.wasm"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	elk ExecutionLayerKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	hashMapStoreKey := sdk.NewKVStoreKey(HashMapStoreKey)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	nicknameStoreKey := sdk.NewKVStoreKey("nickname")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, authtypes.DefaultParamspace)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(nicknameStoreKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(hashMapStoreKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: chainID}, false, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authCapKey, ps, auth.ProtoBaseAccount)
	nicknameKeeper := nickname.NewNicknameKeeper(nicknameStoreKey, cdc, accountKeeper)

	elk := NewExecutionLayerKeeper(cdc, hashMapStoreKey, os.ExpandEnv("$HOME/.casperlabs/.casper-node.sock"),
		accountKeeper, nicknameKeeper)

	gs := types.DefaultGenesisState()
	gs.ChainName = chainID
	gs.Accounts = make([]types.Account, 1)
	gs.Accounts[0] = types.Account{
		Address:             GenesisAccountAddress,
		InitialBalance:      "500000000000000000",
		InitialBondedAmount: "1000000",
	}
	elk.SetGenesisConf(ctx, gs.GenesisConf)
	elk.SetGenesisAccounts(ctx, gs.Accounts)

	return testInput{
		cdc: cdc,
		ctx: ctx,
		elk: elk,
	}
}

func genesis(input testInput) []byte {
	genesisState := types.GenesisState{
		GenesisConf: input.elk.GetGenesisConf(input.ctx),
		Accounts:    input.elk.GetGenesisAccounts(input.ctx),
	}
	genesisConfig, err := types.ToChainSpecGenesisConfig(genesisState)
	if err != nil {
		panic(err)
	}
	response, err := grpc.RunGenesis(input.elk.client, genesisConfig)

	if err != nil {
		panic(err)
	}

	input.elk.SetUnitHashMap(input.ctx, NewUnitHashMap(response.GetSuccess().PoststateHash))

	candidateBlock := input.ctx.CandidateBlock()
	candidateBlock.Hash = []byte{}
	candidateBlock.State = response.GetSuccess().PoststateHash

	systemAccount := make([]byte, 32)
	res, errStr := grpc.Query(input.elk.client, candidateBlock.State, "address", systemAccount, []string{}, genesisConfig.GetProtocolVersion())
	if errStr != "" {
		panic(errStr)
	}

	var storedValue storedvalue.StoredValue
	storedValue, err, _ = storedValue.FromBytes(res)
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

	input.elk.SetProxyContractHash(input.ctx, proxyContractHash)

	return response.GetSuccess().PoststateHash
}

func counterDefine(keeper ExecutionLayerKeeper, parentStateHash []byte) []byte {
	input := setupTestInput()
	proxyHash := keeper.GetProxyContractHash(input.ctx)
	timestamp := time.Now().Unix()
	paymentArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.PaymentMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: "1000000000000000"}}}}}}
	paymentArgsJson, err := DeployArgsToJsonString(paymentArgs)
	if err != nil {
		panic(err)
	}
	cntDefCode := util.LoadWasmFile(path.Join(contractPath, counterDefineWasm))
	protocolVersion := input.elk.MustGetProtocolVersion(input.ctx)
	genesisAddr := input.elk.GetGenesisAccounts(input.ctx)[0]

	deploy, err := util.MakeDeploy(genesisAddr.Address, util.WASM, cntDefCode, "",
		util.WASM, proxyHash, paymentArgsJson, uint64(10), timestamp, input.elk.GetChainName(input.ctx))
	if err != nil {
		panic(fmt.Sprintf("fail to make deploy error : %s", err.Error()))
	}

	deploys := util.MakeInitDeploys()
	deploys = util.AddDeploy(deploys, deploy)

	res, err := grpc.Execute(keeper.client, parentStateHash, timestamp, deploys, &protocolVersion)
	if err != nil {
		panic(fmt.Sprintf("counter define execute: %s", err.Error()))
	}

	postStateHash, _, grpcErr := grpc.Commit(keeper.client, parentStateHash, res.GetSuccess().GetDeployResults()[0].GetExecutionResult().GetEffects().GetTransformMap(), &protocolVersion)
	if grpcErr != "" {
		panic(fmt.Sprintf("counter define commmit: %s", grpcErr))
	}

	return postStateHash

}

func counterCall(keeper ExecutionLayerKeeper, parentStateHash []byte) []byte {
	input := setupTestInput()
	proxyHash := keeper.GetProxyContractHash(input.ctx)
	timestamp := time.Now().Unix()
	paymentArgs := []*consensus.Deploy_Arg{
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_STRING}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_StrValue{
						StrValue: types.PaymentMethodName}}}},
		&consensus.Deploy_Arg{
			Value: &state.CLValueInstance{
				ClType: &state.CLType{Variants: &state.CLType_SimpleType{SimpleType: state.CLType_U512}},
				Value: &state.CLValueInstance_Value{
					Value: &state.CLValueInstance_Value_U512{
						U512: &state.CLValueInstance_U512{
							Value: "1000000000000000"}}}}}}
	paymentArgsJson, err := DeployArgsToJsonString(paymentArgs)
	if err != nil {
		panic(err)
	}
	cntCallCode := util.LoadWasmFile(path.Join(contractPath, counterCallWasm))
	protocolVersion := input.elk.MustGetProtocolVersion(input.ctx)
	genesisAddr := input.elk.GetGenesisAccounts(input.ctx)[0]

	timestamp = time.Now().Unix()
	deploy, err := util.MakeDeploy(genesisAddr.Address, util.WASM, cntCallCode, "",
		util.HASH, proxyHash, paymentArgsJson, uint64(10), timestamp, input.elk.GetChainName(input.ctx))
	if err != nil {
		panic(fmt.Sprintf("fail to make deploy error : %s", err.Error()))
	}
	deploys := util.MakeInitDeploys()
	deploys = util.AddDeploy(deploys, deploy)

	res, err := grpc.Execute(keeper.client, parentStateHash, timestamp, deploys, &protocolVersion)
	if err != nil {
		panic(fmt.Sprintf("counter call execute: %s", err.Error()))
	}

	postStateHash, _, grpcErr := grpc.Commit(keeper.client, parentStateHash, res.GetSuccess().GetDeployResults()[0].GetExecutionResult().GetEffects().GetTransformMap(), &protocolVersion)
	if grpcErr != "" {
		panic(fmt.Sprintf("counter call commit: %s", grpcErr))
	}

	return postStateHash
}
