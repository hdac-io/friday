package types

import (
	"reflect"
	"testing"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/stretchr/testify/require"
)

func TestToProtocolVersion(t *testing.T) {
	// empty string
	got, err := ToProtocolVersion("")
	require.NotNil(t, err)
	require.Nil(t, got)

	// just a number
	got, err = ToProtocolVersion("123")
	require.NotNil(t, err)
	require.Nil(t, got)

	// trailing dot
	got, err = ToProtocolVersion("1.0.0.")
	require.NotNil(t, err)
	require.Nil(t, got)

	// too many digit
	got, err = ToProtocolVersion("1.0.0.0")
	require.NotNil(t, err)
	require.Nil(t, got)

	// valid case
	got, err = ToProtocolVersion("123.456.789")
	require.Nil(t, err)
	expected := state.ProtocolVersion{Major: 123, Minor: 456, Patch: 789}
	if !reflect.DeepEqual(expected, *got) {
		t.Errorf("Bad protocol version parsing. expected %v, got %v", expected, got)
	}
}

func TestToChainSpecGenesisConfig(t *testing.T) {
	// valid input
	genesisState := DefaultGenesisState()
	_, err := ToChainSpecGenesisConfig(genesisState)
	require.Nil(t, err)

	// invalid protocol version
	genesisState.GenesisConf.Genesis.ProtocolVersion = "1.0.0.0"
	_, err = ToChainSpecGenesisConfig(genesisState)
	require.NotNil(t, err)
}
