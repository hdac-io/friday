package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadChainSpec(t *testing.T) {
	got, err := ReadChainSpec("")
	require.NotNil(t, err)
	require.Nil(t, got)

	got, err = ReadChainSpec("../resources/manifest.toml")
	require.Nil(t, err)
	require.NotNil(t, got)

	expected := GenesisConf{
		Genesis: Genesis{
			Name:                "friday-devnet",
			Timestamp:           0,
			MintCodePath:        "mint_install.wasm",
			PosCodePath:         "pos_install.wasm",
			InitialAccountsPath: "accounts.csv",
			ProtocolVersion:     "1.0.0",
		},
		WasmCosts: WasmCosts{
			Regular:           1,
			DivMultiplier:     16,
			MulMultiplier:     4,
			MemMultiplier:     2,
			MemInitialPages:   4096,
			MemGrowPerPage:    8192,
			MemCopyPerByte:    1,
			MaxStackHeight:    65536,
			OpcodesMultiplier: 3,
			OpcodesDivisor:    8,
		},
	}
	if !reflect.DeepEqual(expected, *got) {
		t.Errorf("Bad Unmarshal, expected %v, got %v", expected, *got)
	}
}
