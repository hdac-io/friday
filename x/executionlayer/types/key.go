package types

import (
	"bytes"
	"encoding/binary"

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
	ProtoclVersionKey    = "protocolversion"
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

func GetEEStateKey(height int64) []byte {
	heightBuffer := new(bytes.Buffer)
	binary.Write(heightBuffer, binary.LittleEndian, height)

	return append(EEStateKey, heightBuffer.Bytes()...)
}

func GetValidatorKey(operatorAddr sdk.AccAddress) []byte {
	return append(ValidatorKey, operatorAddr.Bytes()...)
}
