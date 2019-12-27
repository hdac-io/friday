package types

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
