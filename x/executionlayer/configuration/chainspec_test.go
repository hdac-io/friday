package configuration

import (
	"path"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
	"github.com/stretchr/testify/require"
)

const (
	testResourceDir = "../../../tests/resources/executionlayer/genesis"
)

func genesisConfigMock() types.GenesisConf {
	return types.GenesisConf{
		Genesis: types.Genesis{
			Name:            "test-chain",
			Timestamp:       1568805354071,
			MintWasm:        []byte("mint contract bytes"),
			PosWasm:         []byte("pos contract bytes"),
			ProtocolVersion: "1.0.0",
		},
		WasmCosts: types.WasmCosts{
			Regular:           1,
			DivMultiplier:     2,
			MulMultiplier:     3,
			MemMultiplier:     4,
			MemInitialPages:   5,
			MemGrowPerPage:    6,
			MemCopyPerByte:    7,
			MaxStackHeight:    8,
			OpcodesMultiplier: 9,
			OpcodesDivisor:    10,
		},
	}
}

func TestParseGenesisChainSpecBasic(t *testing.T) {
	// valid input
	got, err := ParseGenesisChainSpec(path.Join(testResourceDir, "manifest.toml"))
	require.Nil(t, err)
	expected := genesisConfigMock()
	require.Equal(t, expected.Genesis.Name, got.Genesis.Name)
	require.Equal(t, expected.Genesis.Timestamp, got.Genesis.Timestamp)
	require.Equal(t, expected.Genesis.MintWasm, got.Genesis.MintWasm)
	require.Equal(t, expected.Genesis.PosWasm, got.Genesis.PosWasm)
	require.Equal(t, expected.Genesis.ProtocolVersion, got.Genesis.ProtocolVersion)

	if !reflect.DeepEqual(expected.WasmCosts, got.WasmCosts) {
		t.Errorf("Bad WasmCosts, expected %v, got %v", expected.WasmCosts, got.WasmCosts)
	}
}

func TestParseGenesisChainSpecWithMissingField(t *testing.T) {
	_, err := ParseGenesisChainSpec(path.Join(testResourceDir, "genesis_with_missing_field.toml"))
	require.NotNil(t, err)

	elError, ok := err.(sdk.Error)
	if !ok || elError.Code() != types.CodeTomlParse {
		t.Errorf("Should fail with ErrTomlParse. Code: %v", err)
	}
}

func TestParseGenesisChainSpecWithInvalidWasmPath(t *testing.T) {
	_, err := ParseGenesisChainSpec(
		path.Join(testResourceDir, "genesis_with_invalid_wasm_path.toml"))
	require.NotNil(t, err)

	if !strings.Contains(err.Error(), "invalid_path.wasm") {
		t.Errorf("Should fail with invalid wasm path but fail with: %v", err)
	}
}
