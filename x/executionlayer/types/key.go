package types

import (
	sdk "github.com/hdac-io/friday/types"
)

const (
	// ModuleName uses for schema name in key-value store
	ModuleName = "executionlayer"

	// StoreKey sets schema name from ModuleName
	HashMapStoreKey = ModuleName + "_hashmap"

	// key value
	GenesisBlockHashKey = "genesisblockhash"
	GenesisConfigKey    = "genesisconf"
	GenesisAccountKey   = "genesisaccount"
	CandidateBlockKey   = "candidateblock"
)

var (
	ValidatorKey = []byte{0x21}
)

func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorKey, operatorAddr.Bytes()...)
}
