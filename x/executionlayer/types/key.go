package types

import (
	sdk "github.com/hdac-io/friday/types"
)

const (
	// ModuleName uses for schema name in key-value store
	ModuleName = "contract"

	// StoreKey sets schema name from ModuleName
	HashMapStoreKey = ModuleName + "_hashmap"

	// key value
	GenesisBlockHashKey  = "genesisblockhash"
	GenesisConfigKey     = "genesisconf"
	GenesisAccountKey    = "genesisaccount"
	CandidateBlockKey    = "candidateblock"
	ProxyContractHashKey = "proxycontractkey"
)

var (
	EEStateKey   = []byte{0x11}
	ValidatorKey = []byte{0x21}
)

type (
	ContractAddress     = sdk.ContractAddress
	ContractHashAddress = sdk.ContractHashAddress
	ContractUrefAddress = sdk.ContractUrefAddress
)

func GetEEStateKey(eeState []byte) []byte {
	return append(EEStateKey, eeState...)
}

func GetValidatorKey(operatorAddr sdk.AccAddress) []byte {
	return append(ValidatorKey, operatorAddr.Bytes()...)
}
