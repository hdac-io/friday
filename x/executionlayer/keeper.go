package executionlayer

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"

	"github.com/hdac-io/tendermint/crypto"
	secp256k1 "github.com/hdac-io/tendermint/crypto/secp256k1"

	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/hdac-io/friday/x/readablename"
)

type ExecutionLayerKeeper struct {
	HashMapStoreKey    sdk.StoreKey
	client             ipc.ExecutionEngineServiceClient
	AccountKeeper      auth.AccountKeeper
	ReadableNameKeeper readablename.ReadableNameKeeper
	cdc                *codec.Codec
}

func NewExecutionLayerKeeper(
	cdc *codec.Codec, hashMapStoreKey sdk.StoreKey, path string,
	accountKeeper auth.AccountKeeper,
	reaablenameKeeper readablename.ReadableNameKeeper) ExecutionLayerKeeper {

	return ExecutionLayerKeeper{
		HashMapStoreKey:    hashMapStoreKey,
		client:             grpc.Connect(path),
		AccountKeeper:      accountKeeper,
		ReadableNameKeeper: reaablenameKeeper,
		cdc:                cdc,
	}
}

func (k ExecutionLayerKeeper) MustGetProtocolVersion(ctx sdk.Context) state.ProtocolVersion {
	genesisConf := k.GetGenesisConf(ctx)
	pv, err := types.ToProtocolVersion(genesisConf.Genesis.ProtocolVersion)
	if err != nil {
		panic(fmt.Errorf("System has invalid protocol version: %v", err))
	}
	return *pv
}

// -----------------------------------------------------------------------------------------------------------

// SetUnitHashMap map unitHash to blockHash
func (k ExecutionLayerKeeper) SetUnitHashMap(ctx sdk.Context, blockHash []byte, unitHash UnitHashMap) bool {
	if k.isEmptyHash(blockHash) {
		blockHash = []byte(types.GenesisBlockHashKey)
	}
	if k.isEmptyHash(unitHash.EEState) || len(unitHash.EEState) != 32 {
		return false
	}

	unitBytes, err := k.cdc.MarshalBinaryBare(unitHash)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set(types.GetEEStateKey(blockHash), unitBytes)

	return true
}

// GetUnitHashMap returns a UnitHashMap for blockHash
func (k ExecutionLayerKeeper) GetUnitHashMap(ctx sdk.Context, blockHash []byte) UnitHashMap {
	if k.isEmptyHash(blockHash) {
		blockHash = []byte(types.GenesisBlockHashKey)
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	unitBytes := store.Get(types.GetEEStateKey(blockHash))
	var unit UnitHashMap
	k.cdc.UnmarshalBinaryBare(unitBytes, &unit)
	return unit
}

// SetEEState map eeState to blockHash
func (k ExecutionLayerKeeper) SetEEState(ctx sdk.Context, blockHash []byte, eeState []byte) bool {
	if k.isEmptyHash(blockHash) {
		blockHash = []byte(types.GenesisBlockHashKey)
	}
	if k.isEmptyHash(eeState) || len(eeState) != 32 {
		return false
	}

	unit := UnitHashMap{
		EEState: eeState,
	}

	return k.SetUnitHashMap(ctx, blockHash, unit)
}

// GetEEState returns a eeState for blockHash
func (k ExecutionLayerKeeper) GetEEState(ctx sdk.Context, blockHash []byte) []byte {
	if k.isEmptyHash(blockHash) {
		blockHash = []byte(types.GenesisBlockHashKey)
	}
	unit := k.GetUnitHashMap(ctx, blockHash)
	return unit.EEState
}

// Transfer function executes "Execute" of Execution layer, that is specialized for transfer
// Difference of general execution
//   1) Raw account is needed for checking address existence
//   2) Fixed transfer & payemtn WASMs are needed
func (k ExecutionLayerKeeper) Transfer(
	ctx sdk.Context,
	tokenContractAddress string,
	fromPubkey, toPubkey secp256k1.PubKeySecp256k1,
	transferCode []byte,
	transferAbi []byte,
	paymentCode []byte,
	paymentAbi []byte,
	gasPrice uint64) error {

	k.SetAccountIfNotExists(ctx, toPubkey)
	err := k.Execute(ctx, k.GetCandidateBlockHash(ctx), fromPubkey, tokenContractAddress, transferCode, transferAbi, paymentCode, paymentAbi, gasPrice)
	if err != nil {
		return err
	}

	return nil
}

// Execute is general execution
func (k ExecutionLayerKeeper) Execute(ctx sdk.Context,
	blockHash []byte,
	execPubkey secp256k1.PubKeySecp256k1,
	contractAddress string,
	sessionCode []byte,
	sessionArgs []byte,
	paymentCode []byte,
	paymentArgs []byte,
	gasPrice uint64) error {

	copiedBlockhash := blockHash
	if bytes.Equal(copiedBlockhash, []byte{0}) {
		copiedBlockhash = k.GetCandidateBlockHash(ctx)
	}

	// Parameter preparation
	unitHash := k.GetUnitHashMap(ctx, copiedBlockhash)
	protocolVersion := k.MustGetProtocolVersion(ctx)

	exexAddr := sdk.GetEEAddressFromSecp256k1PubKey(execPubkey)

	// Execute
	deploys := []*ipc.DeployItem{}
	deploy := util.MakeDeploy(exexAddr.Bytes(), sessionCode, sessionArgs, paymentCode, paymentArgs, gasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
	deploys = append(deploys, deploy)
	reqExecute := &ipc.ExecuteRequest{
		ParentStateHash: unitHash.EEState,
		BlockTime:       uint64(ctx.BlockTime().Unix()),
		Deploys:         deploys,
		ProtocolVersion: &protocolVersion,
	}
	resExecute, err := k.client.Execute(ctx.Context(), reqExecute)
	if err != nil {
		return err
	}

	effects := []*transforms.TransformEntry{}
	switch resExecute.GetResult().(type) {
	case *ipc.ExecuteResponse_Success:
		for _, res := range resExecute.GetSuccess().GetDeployResults() {
			switch res.GetExecutionResult().GetError().GetValue().(type) {
			case *ipc.DeployError_GasError:
				err = types.ErrGRpcExecuteDeployGasError(types.DefaultCodespace)
			case *ipc.DeployError_ExecError:
				err = types.ErrGRpcExecuteDeployExecError(types.DefaultCodespace, res.GetExecutionResult().GetError().GetExecError().GetMessage())
			default:
				effects = append(effects, res.GetExecutionResult().GetEffects().GetTransformMap()...)
			}

		}
	case *ipc.ExecuteResponse_MissingParent:
		err = types.ErrGRpcExecuteMissingParent(types.DefaultCodespace, util.EncodeToHexString(resExecute.GetMissingParent().GetHash()))
	}
	if err != nil {
		return err
	}

	// Commit
	postStateHash, bonds, errGrpc := grpc.Commit(k.client, unitHash.EEState, effects, &protocolVersion)
	if errGrpc != "" {
		return fmt.Errorf(errGrpc)
	}

	k.SetEEState(ctx, copiedBlockhash, postStateHash)
	k.SetCandidateBlockBond(ctx, bonds)

	return nil
}

// GetQueryResult queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryResult(ctx sdk.Context,
	blockhash []byte, keyType string, keyData string, path string) (state.Value, error) {
	arrPath := strings.Split(path, "/")

	protocolVersion := k.MustGetProtocolVersion(ctx)
	unitHash := k.GetUnitHashMap(ctx, blockhash)
	keyDataBytes, err := toBytes(keyType, keyData, k.ReadableNameKeeper, ctx)
	if err != nil {
		return state.Value{}, err
	}
	res, errstr := grpc.Query(k.client, unitHash.EEState, keyType, keyDataBytes, arrPath, &protocolVersion)
	if errstr != "" {
		return state.Value{}, fmt.Errorf(errstr)
	}

	return *res, nil
}

// GetQueryResultSimple queries without state hash.
// State hash comes from Tendermint block state - EE state mapping DB
func (k ExecutionLayerKeeper) GetQueryResultSimple(ctx sdk.Context,
	keyType string, keyData string, path string) (state.Value, error) {
	currBlock := k.GetCandidateBlockHash(ctx)
	res, err := k.GetQueryResult(ctx, currBlock, keyType, keyData, path)
	if err != nil {
		return state.Value{}, err
	}

	return res, nil
}

// GetQueryBalanceResult queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResult(ctx sdk.Context, blockhash []byte, pubkey secp256k1.PubKeySecp256k1) (string, error) {
	unitHash := k.GetUnitHashMap(ctx, blockhash)
	protocolVersion := k.MustGetProtocolVersion(ctx)
	addr, err := sdk.GetEEAddressFromCryptoPubkey(pubkey)
	if err != nil {
		return "", err
	}

	res, grpcErr := grpc.QueryBalance(k.client, unitHash.EEState, addr.Bytes(), &protocolVersion)
	if grpcErr != "" {
		return "", fmt.Errorf(grpcErr)
	}

	return res, nil
}

// GetQueryBalanceResultSimple queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResultSimple(ctx sdk.Context, pubkey secp256k1.PubKeySecp256k1) (string, error) {
	res, err := k.GetQueryBalanceResult(ctx, k.GetCandidateBlockHash(ctx), pubkey)
	if err != nil {
		return "", err
	}

	return res, nil
}

// -----------------------------------------------------------------------------------------------------------

// GetGenesisConf retrieves GenesisConf from sdk store
func (k ExecutionLayerKeeper) GetGenesisConf(ctx sdk.Context) types.GenesisConf {
	store := ctx.KVStore(k.HashMapStoreKey)
	genesisConfBytes := store.Get([]byte(types.GenesisConfigKey))

	var genesisConf types.GenesisConf
	k.cdc.UnmarshalBinaryBare(genesisConfBytes, &genesisConf)
	return genesisConf
}

// SetGenesisConf saves GenesisConf in sdk store
func (k ExecutionLayerKeeper) SetGenesisConf(ctx sdk.Context, genesisConf types.GenesisConf) {
	store := ctx.KVStore(k.HashMapStoreKey)
	genesisConfBytes := k.cdc.MustMarshalBinaryBare(genesisConf)
	store.Set([]byte(types.GenesisConfigKey), genesisConfBytes)
}

// GetGenesisAccounts retrieves GenesisAccounts in sdk store
func (k ExecutionLayerKeeper) GetGenesisAccounts(ctx sdk.Context) []types.Account {
	store := ctx.KVStore(k.HashMapStoreKey)
	genesisAccountsBytes := store.Get([]byte(types.GenesisAccountKey))
	if genesisAccountsBytes == nil {
		return nil
	}
	var genesisAccounts []types.Account
	k.cdc.UnmarshalBinaryBare(genesisAccountsBytes, &genesisAccounts)
	return genesisAccounts
}

// SetGenesisAccounts saves GenesisAccounts in sdk store
func (k ExecutionLayerKeeper) SetGenesisAccounts(ctx sdk.Context, accounts []types.Account) {
	if accounts == nil {
		panic(fmt.Errorf("Nil is not allowed for GenesisAccounts"))
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	genesisAccountsBytes := k.cdc.MustMarshalBinaryBare(accounts)
	store.Set([]byte(types.GenesisAccountKey), genesisAccountsBytes)
}

// GetChainName retrieves ChainName in sdk store
func (k ExecutionLayerKeeper) GetChainName(ctx sdk.Context) string {
	store := ctx.KVStore(k.HashMapStoreKey)
	chainNameBytes := store.Get([]byte("chainname"))
	return string(chainNameBytes)
}

// SetChainName saves ChainName in sdk store
func (k ExecutionLayerKeeper) SetChainName(ctx sdk.Context, chainName string) {
	if chainName == "" {
		panic(fmt.Errorf("Empty string is not allowed for ChainName"))
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set([]byte("chainname"), []byte(chainName))
}

// SetAccountIfNotExists runs if network has no given account
func (k ExecutionLayerKeeper) SetAccountIfNotExists(ctx sdk.Context, pubkey secp256k1.PubKeySecp256k1) {
	// Recepient account existence check, if not, create one
	account := sdk.AccAddress(pubkey.Address())
	toAddressAccountObject := k.AccountKeeper.GetAccount(ctx, account)
	if toAddressAccountObject == nil {
		toAddressAccountObject = k.AccountKeeper.NewAccountWithAddress(ctx, account)
		k.AccountKeeper.SetAccount(ctx, toAddressAccountObject)
	}
}

// -----------------------------------------------------------------------------------------------------------
// GetCandidateBlock returns current block hash
func (k ExecutionLayerKeeper) GetCandidateBlock(ctx sdk.Context) types.CandidateBlock {
	store := ctx.KVStore(k.HashMapStoreKey)
	candidateBlockBytes := store.Get([]byte(types.CandidateBlockKey))
	var candidateBlock types.CandidateBlock
	k.cdc.UnmarshalBinaryBare(candidateBlockBytes, &candidateBlock)

	return candidateBlock
}

func (k ExecutionLayerKeeper) SetCandidateBlock(ctx sdk.Context, candidateBlock types.CandidateBlock) {
	store := ctx.KVStore(k.HashMapStoreKey)

	// It stores the bonds received from the execution-engine.
	//Even though they are executed in the same order for each node,
	//the order of the data is different and the state of the keeper is different.
	//This is an sort to reflect this.
	sort.Slice(candidateBlock.Bonds, func(i, j int) bool {
		return bytes.Compare(candidateBlock.Bonds[i].GetValidatorPublicKey(), candidateBlock.Bonds[j].GetValidatorPublicKey()) > 0
	})
	candidateBlockBytes := k.cdc.MustMarshalBinaryBare(candidateBlock)
	store.Set([]byte(types.CandidateBlockKey), candidateBlockBytes)
}

// GetCandidateBlockHash returns current block hash
func (k ExecutionLayerKeeper) GetCandidateBlockHash(ctx sdk.Context) []byte {
	candidateBlock := k.GetCandidateBlock(ctx)

	return candidateBlock.Hash
}

// SetCandidateBlockHash saves current block hash
func (k ExecutionLayerKeeper) SetCandidateBlockHash(ctx sdk.Context, blockHash []byte) {
	candidateBlock := k.GetCandidateBlock(ctx)
	candidateBlock.Hash = blockHash
	k.SetCandidateBlock(ctx, candidateBlock)
}

func (k ExecutionLayerKeeper) GetCandidateBlockBond(ctx sdk.Context) []*ipc.Bond {
	candidateBlock := k.GetCandidateBlock(ctx)
	return candidateBlock.Bonds
}

func (k ExecutionLayerKeeper) SetCandidateBlockBond(ctx sdk.Context, bonds []*ipc.Bond) {
	candidateBlock := k.GetCandidateBlock(ctx)
	candidateBlock.Bonds = bonds
	k.SetCandidateBlock(ctx, candidateBlock)
}

// -----------------------------------------------------------------------------------------------------------

func (k ExecutionLayerKeeper) GetValidator(ctx sdk.Context, eeAddress sdk.EEAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.HashMapStoreKey)
	validatorBytes := store.Get(types.GetValidatorKey(eeAddress))
	if validatorBytes == nil {
		return validator, false
	}
	validator = types.MustUnmarshalValidator(k.cdc, validatorBytes)

	return validator, true
}

func (k ExecutionLayerKeeper) SetValidator(ctx sdk.Context, eeAddress sdk.EEAddress, validator types.Validator) {
	store := ctx.KVStore(k.HashMapStoreKey)
	validatorBytes := types.MustMarshalValidator(k.cdc, validator)
	store.Set(types.GetValidatorKey(eeAddress), validatorBytes)
}

func (k ExecutionLayerKeeper) GetValidatorConsPubKey(ctx sdk.Context, eeAddress sdk.EEAddress) crypto.PubKey {
	validator, _ := k.GetValidator(ctx, eeAddress)

	return validator.ConsPubKey
}

func (k ExecutionLayerKeeper) SetValidatorConsPubKey(ctx sdk.Context, eeAddress sdk.EEAddress, pubKey crypto.PubKey) {
	validator, _ := k.GetValidator(ctx, eeAddress)
	validator.ConsPubKey = pubKey
	k.SetValidator(ctx, eeAddress, validator)
}

func (k ExecutionLayerKeeper) GetValidatorDescription(ctx sdk.Context, eeAddress sdk.EEAddress) types.Description {
	validator, _ := k.GetValidator(ctx, eeAddress)

	return validator.Description
}

func (k ExecutionLayerKeeper) SetValidatorDescription(ctx sdk.Context, eeAddress sdk.EEAddress, description types.Description) {
	validator, _ := k.GetValidator(ctx, eeAddress)
	validator.Description = description
	k.SetValidator(ctx, eeAddress, validator)
}

func (k ExecutionLayerKeeper) GetValidatorStake(ctx sdk.Context, eeAddress sdk.EEAddress) string {
	validator, _ := k.GetValidator(ctx, eeAddress)

	return validator.Stake
}

func (k ExecutionLayerKeeper) SetValidatorStake(ctx sdk.Context, eeAddress sdk.EEAddress, stake string) {
	validator, _ := k.GetValidator(ctx, eeAddress)
	validator.Stake = stake
	k.SetValidator(ctx, eeAddress, validator)
}

func (k ExecutionLayerKeeper) GetAllValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.HashMapStoreKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validators = append(validators, validator)
	}
	return validators
}

// -----------------------------------------------------------------------------------------------------------

func (k ExecutionLayerKeeper) isEmptyHash(src []byte) bool {
	return bytes.Equal([]byte{}, src)
}
