package executionlayer

import (
	"fmt"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"

	"github.com/hdac-io/tendermint/crypto"

	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/auth"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/hdac-io/friday/x/nickname"
)

type ExecutionLayerKeeper struct {
	HashMapStoreKey sdk.StoreKey
	client          ipc.ExecutionEngineServiceClient
	AccountKeeper   auth.AccountKeeper
	NicknameKeeper  nickname.NicknameKeeper
	cdc             *codec.Codec
}

func NewExecutionLayerKeeper(
	cdc *codec.Codec, hashMapStoreKey sdk.StoreKey, path string,
	accountKeeper auth.AccountKeeper,
	nicknameKeeper nickname.NicknameKeeper) ExecutionLayerKeeper {

	return ExecutionLayerKeeper{
		HashMapStoreKey: hashMapStoreKey,
		client:          grpc.Connect(path),
		AccountKeeper:   accountKeeper,
		NicknameKeeper:  nicknameKeeper,
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
func (k ExecutionLayerKeeper) SetUnitHashMap(ctx sdk.Context, unitHash UnitHashMap) bool {
	if len(unitHash.EEState) == 0 || len(unitHash.EEState) != 32 {
		return false
	}

	unitBytes, err := k.cdc.MarshalBinaryBare(unitHash)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set(types.GetEEStateKey(ctx.BlockHeight()), unitBytes)

	return true
}

// GetUnitHashMap returns a UnitHashMap for blockHash
func (k ExecutionLayerKeeper) GetUnitHashMap(ctx sdk.Context, height int64) UnitHashMap {
	store := ctx.KVStore(k.HashMapStoreKey)
	unitBytes := store.Get(types.GetEEStateKey(height))
	var unit UnitHashMap
	k.cdc.UnmarshalBinaryBare(unitBytes, &unit)
	return unit
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
func (k ExecutionLayerKeeper) SetAccountIfNotExists(ctx sdk.Context, addr sdk.AccAddress) {
	// Recepient account existence check, if not, create one
	toAddressAccountObject := k.AccountKeeper.GetAccount(ctx, addr)
	if toAddressAccountObject == nil {
		toAddressAccountObject = k.AccountKeeper.NewAccountWithAddress(ctx, addr)
		k.AccountKeeper.SetAccount(ctx, toAddressAccountObject)
	}
}

// -----------------------------------------------------------------------------------------------------------

func (k ExecutionLayerKeeper) GetValidator(ctx sdk.Context, accAddress sdk.AccAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.HashMapStoreKey)
	validatorBytes := store.Get(types.GetValidatorKey(accAddress))
	if validatorBytes == nil {
		return validator, false
	}
	validator = types.MustUnmarshalValidator(k.cdc, validatorBytes)

	return validator, true
}

func (k ExecutionLayerKeeper) SetValidator(ctx sdk.Context, accAddress sdk.AccAddress, validator types.Validator) {
	store := ctx.KVStore(k.HashMapStoreKey)
	validatorBytes := types.MustMarshalValidator(k.cdc, validator)
	store.Set(types.GetValidatorKey(accAddress), validatorBytes)
}

func (k ExecutionLayerKeeper) GetValidatorConsPubKey(ctx sdk.Context, accAddress sdk.AccAddress) crypto.PubKey {
	validator, _ := k.GetValidator(ctx, accAddress)

	return validator.ConsPubKey
}

func (k ExecutionLayerKeeper) SetValidatorConsPubKey(ctx sdk.Context, accAddress sdk.AccAddress, pubKey crypto.PubKey) {
	validator, _ := k.GetValidator(ctx, accAddress)
	validator.ConsPubKey = pubKey
	k.SetValidator(ctx, accAddress, validator)
}

func (k ExecutionLayerKeeper) GetValidatorDescription(ctx sdk.Context, accAddress sdk.AccAddress) types.Description {
	validator, _ := k.GetValidator(ctx, accAddress)

	return validator.Description
}

func (k ExecutionLayerKeeper) SetValidatorDescription(ctx sdk.Context, accAddress sdk.AccAddress, description types.Description) {
	validator, _ := k.GetValidator(ctx, accAddress)
	validator.Description = description
	k.SetValidator(ctx, accAddress, validator)
}

func (k ExecutionLayerKeeper) GetValidatorStake(ctx sdk.Context, accAddress sdk.AccAddress) string {
	validator, _ := k.GetValidator(ctx, accAddress)

	return validator.Stake
}

func (k ExecutionLayerKeeper) SetValidatorStake(ctx sdk.Context, accAddress sdk.AccAddress, stake string) {
	validator, _ := k.GetValidator(ctx, accAddress)
	validator.Stake = stake
	k.SetValidator(ctx, accAddress, validator)
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

// GetProxyContractHash retrieves proxy_contract_hash
func (k ExecutionLayerKeeper) GetProxyContractHash(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.HashMapStoreKey)
	proxyContractHash := store.Get([]byte(types.ProxyContractHashKey))

	return proxyContractHash
}

// SetProxyContractHash save proxy_contract_hash value.
func (k ExecutionLayerKeeper) SetProxyContractHash(ctx sdk.Context, contractHash []byte) {
	store := ctx.KVStore(k.HashMapStoreKey)
	store.Set([]byte(types.ProxyContractHashKey), contractHash)
}
