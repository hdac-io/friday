package executionlayer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"

	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/tendermint/crypto"

	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

type ExecutionLayerKeeper struct {
	HashMapStoreKey sdk.StoreKey
	client          ipc.ExecutionEngineServiceClient
	AccountKeeper   auth.AccountKeeper
	cdc             *codec.Codec
}

func NewExecutionLayerKeeper(
	cdc *codec.Codec, hashMapStoreKey sdk.StoreKey, path string, accountKeeper auth.AccountKeeper) ExecutionLayerKeeper {

	return ExecutionLayerKeeper{
		HashMapStoreKey: hashMapStoreKey,
		client:          grpc.Connect(path),
		AccountKeeper:   accountKeeper,
		cdc:             cdc,
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
	store.Set(blockHash, unitBytes)

	return true
}

// GetUnitHashMap returns a UnitHashMap for blockHash
func (k ExecutionLayerKeeper) GetUnitHashMap(ctx sdk.Context, blockHash []byte) UnitHashMap {
	if k.isEmptyHash(blockHash) {
		blockHash = []byte(types.GenesisBlockHashKey)
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	unitBytes := store.Get(blockHash)
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

	unitBytes, err := k.cdc.MarshalBinaryBare(unit)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set(blockHash, unitBytes)

	return true
}

// GetEEState returns a eeState for blockHash
func (k ExecutionLayerKeeper) GetEEState(ctx sdk.Context, blockHash []byte) []byte {
	if k.isEmptyHash(blockHash) {
		blockHash = []byte(types.GenesisBlockHashKey)
	}
	store := ctx.KVStore(k.HashMapStoreKey)
	unitBytes := store.Get(blockHash)
	var unit UnitHashMap
	k.cdc.UnmarshalBinaryBare(unitBytes, &unit)
	return unit.EEState
}

// Transfer function executes "Execute" of Execution layer, that is specialized for transfer
// Difference of general execution
//   1) Raw account is needed for checking address existence
//   2) Fixed transfer & payemtn WASMs are needed
func (k ExecutionLayerKeeper) Transfer(
	ctx sdk.Context,
	tokenOwnerAccount, fromAddress, toAddress sdk.AccAddress,
	transferCode []byte,
	transferAbi []byte,
	paymentCode []byte,
	paymentAbi []byte,
	gasPrice uint64) error {

	k.SetAccountIfNotExists(ctx, toAddress)
	err := k.Execute(ctx, k.GetCandidateBlockHash(ctx), fromAddress, tokenOwnerAccount, transferCode, transferAbi, paymentCode, paymentAbi, gasPrice)
	if err != nil {
		return err
	}

	return nil
}

// Execute is general execution
func (k ExecutionLayerKeeper) Execute(ctx sdk.Context,
	blockHash []byte,
	execAccount sdk.AccAddress,
	contractOwnerAccount sdk.AccAddress,
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
	execAccountPubKey := types.ToPublicKey(execAccount)
	unitHash := k.GetUnitHashMap(ctx, copiedBlockhash)
	protocolVersion := k.MustGetProtocolVersion(ctx)

	// Execute
	deploys := util.MakeInitDeploys()
	deploy := util.MakeDeploy(execAccountPubKey, sessionCode, sessionArgs, paymentCode, paymentArgs, gasPrice, ctx.BlockTime().Unix(), ctx.ChainID())
	deploys = util.AddDeploy(deploys, deploy)
	effects, errGrpc := grpc.Execute(k.client, unitHash.EEState, ctx.BlockTime().Unix(), deploys, &protocolVersion)
	if errGrpc != "" {
		return fmt.Errorf(errGrpc)
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
	keyDataBytes, err := toBytes(keyType, keyData)
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
func (k ExecutionLayerKeeper) GetQueryBalanceResult(ctx sdk.Context, blockhash []byte, address types.PublicKey) (string, error) {
	unitHash := k.GetUnitHashMap(ctx, blockhash)
	protocolVersion := k.MustGetProtocolVersion(ctx)
	res, err := grpc.QueryBalance(k.client, unitHash.EEState, address, &protocolVersion)
	if err != "" {
		return "", fmt.Errorf(err)
	}

	return res, nil
}

// GetQueryBalanceResultSimple queries with whole parameters
func (k ExecutionLayerKeeper) GetQueryBalanceResultSimple(ctx sdk.Context, address types.PublicKey) (string, error) {
	res, err := k.GetQueryBalanceResult(ctx, k.GetCandidateBlockHash(ctx), address)
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
func (k ExecutionLayerKeeper) SetAccountIfNotExists(ctx sdk.Context, account sdk.AccAddress) {
	// Recepient account existence check, if not, create one
	toAddressAccountObject := k.AccountKeeper.GetAccount(ctx, account)
	if toAddressAccountObject == nil {
		toAddressAccountObject = k.AccountKeeper.NewAccountWithAddress(ctx, account)
		k.AccountKeeper.SetAccount(ctx, toAddressAccountObject)
	}
}

// -----------------------------------------------------------------------------------------------------------
// GetCurrentBlockHash returns current block hash
func (k ExecutionLayerKeeper) GetCandidateBlock(ctx sdk.Context) types.CandidateBlock {
	store := ctx.KVStore(k.HashMapStoreKey)
	candidateBlockBytes := store.Get([]byte(types.CandidateBlockKey))
	var candidateBlock types.CandidateBlock
	k.cdc.UnmarshalBinaryBare(candidateBlockBytes, &candidateBlock)

	return candidateBlock
}

func (k ExecutionLayerKeeper) SetCandidateBlock(ctx sdk.Context, candidateBlock types.CandidateBlock) {
	store := ctx.KVStore(k.HashMapStoreKey)
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

func (k ExecutionLayerKeeper) GetValidator(ctx sdk.Context, accAddress []byte) types.Validator {
	store := ctx.KVStore(k.HashMapStoreKey)
	validatorBytes := store.Get(accAddress)
	var validator types.Validator
	k.cdc.UnmarshalBinaryBare(validatorBytes, &validator)

	return validator
}

func (k ExecutionLayerKeeper) SetValidator(ctx sdk.Context, accAddress []byte, validator types.Validator) {
	store := ctx.KVStore(k.HashMapStoreKey)
	validatorBytes := k.cdc.MustMarshalBinaryBare(validator)
	store.Set(accAddress, validatorBytes)
}

func (k ExecutionLayerKeeper) GetValidatorOperatorAddress(ctx sdk.Context, accAddress []byte) sdk.ValAddress {
	validator := k.GetValidator(ctx, accAddress)

	return validator.OperatorAddress
}

func (k ExecutionLayerKeeper) SetValidatorOperatorAddress(ctx sdk.Context, accAddress []byte, valAddress sdk.ValAddress) {
	validator := k.GetValidator(ctx, accAddress)
	validator.OperatorAddress = valAddress
	k.SetValidator(ctx, accAddress, validator)
}

func (k ExecutionLayerKeeper) GetValidatorConsPubKey(ctx sdk.Context, accAddress []byte) crypto.PubKey {
	validator := k.GetValidator(ctx, accAddress)

	return validator.ConsPubKey
}

func (k ExecutionLayerKeeper) SetValidatorConsPubKey(ctx sdk.Context, accAddress []byte, pubKey crypto.PubKey) {
	validator := k.GetValidator(ctx, accAddress)
	validator.ConsPubKey = pubKey
	k.SetValidator(ctx, accAddress, validator)
}

func (k ExecutionLayerKeeper) GetValidatorDescription(ctx sdk.Context, accAddress []byte) types.Description {
	validator := k.GetValidator(ctx, accAddress)

	return validator.Description
}

func (k ExecutionLayerKeeper) SetValidatorDescription(ctx sdk.Context, accAddress []byte, description types.Description) {
	validator := k.GetValidator(ctx, accAddress)
	validator.Description = description
	k.SetValidator(ctx, accAddress, validator)
}

// -----------------------------------------------------------------------------------------------------------
func (k ExecutionLayerKeeper) isEmptyHash(src []byte) bool {
	return bytes.Equal([]byte{}, src)
}
