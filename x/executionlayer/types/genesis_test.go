package types

import (
	"os"
	"reflect"
	"testing"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/stretchr/testify/require"
)

const (
	mintCodePath = "$HOME/.nodef/contracts/mint_install.wasm"
	posCodePath  = "$HOME/.nodef/contracts/pos_install.wasm"
)

func TestToProtocolVersion(t *testing.T) {
	// empty string
	got, err := toProtocolVersion("")
	require.NotNil(t, err)
	require.Nil(t, got)

	// just a number
	got, err = toProtocolVersion("123")
	require.NotNil(t, err)
	require.Nil(t, got)

	// trailing dot
	got, err = toProtocolVersion("1.0.0.")
	require.NotNil(t, err)
	require.Nil(t, got)

	// too many digit
	got, err = toProtocolVersion("1.0.0.0")
	require.NotNil(t, err)
	require.Nil(t, got)

	// valid case
	got, err = toProtocolVersion("123.456.789")
	require.Nil(t, err)
	expected := state.ProtocolVersion{Major: 123, Minor: 456, Patch: 789}
	if !reflect.DeepEqual(expected, *got) {
		t.Errorf("Bad protocol version parsing. expected %v, got %v", expected, got)
	}
}

func TestToChainSpecGenesisConfig(t *testing.T) {
	// valid input
	genesisState := DefaultGenesisState()
	genesisState.GenesisConf.Genesis.MintCodePath = os.ExpandEnv(mintCodePath)
	genesisState.GenesisConf.Genesis.PosCodePath = os.ExpandEnv(posCodePath)
	_, err := ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.Nil(t, err)

	// invalid system contract path
	genesisState.GenesisConf.Genesis.MintCodePath = "test-odd-path"
	_, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.NotNil(t, err)

	genesisState.GenesisConf.Genesis.MintCodePath = os.ExpandEnv(mintCodePath)
	genesisState.GenesisConf.Genesis.PosCodePath = "test-odd-path"
	_, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.NotNil(t, err)

	// revert to valid input
	genesisState.GenesisConf.Genesis.MintCodePath = os.ExpandEnv(mintCodePath)
	genesisState.GenesisConf.Genesis.PosCodePath = os.ExpandEnv(posCodePath)

	// invalid protocol version
	genesisState.GenesisConf.Genesis.ProtocolVersion = "1.0.0.0"
	_, err = ToChainSpecGenesisConfig(genesisState.GenesisConf)
	require.NotNil(t, err)
}
