package types

import (
	"encoding/base64"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/stretchr/testify/require"
)

func TestReadChainSpec(t *testing.T) {
	got, err := readChainSpec("")
	require.NotNil(t, err)
	require.Nil(t, got)

	got, err = readChainSpec("../resources/manifest.toml")
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

func TestReadGenesisAccountsCsv(t *testing.T) {
	got, err := readGenesisAccountsCsv("")
	require.NotNil(t, err)
	require.Empty(t, got)

	got, err = readGenesisAccountsCsv("../resources/accounts.csv")
	require.Nil(t, err)
	require.Equal(t, 1, len(got))

	expected := Account{
		publicKey:           "s8qP7TauBe0WoHUDEKyFR99XM6q7aGzacLa6M6vHtO0=",
		initialBalance:      "50000000000",
		initialBondedAmount: "1000000",
	}
	if !reflect.DeepEqual(expected, got[0]) {
		t.Errorf("Bad accounts.csv, expected %v, got %v", expected, got)
	}
}

func TestFromAccount(t *testing.T) {

}

func TestParseProtocolVersion(t *testing.T) {

}

func genesisConfigMock() (*ipc.ChainSpec_GenesisConfig, error) {
	expected := ipc.ChainSpec_GenesisConfig{}
	expected.Name = "friday-devnet"
	expected.Timestamp = 0
	expected.ProtocolVersion = &state.ProtocolVersion{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}

	// load mint_install.wasm, pos_install.wasm
	var err error
	expected.MintInstaller, err = ioutil.ReadFile("../resources/mint_install.wasm")
	if err != nil {
		return nil, err
	}
	expected.PosInstaller, err = ioutil.ReadFile("../resources/pos_install.wasm")
	if err != nil {
		return nil, err
	}

	// GenesisAccount
	accounts := make([]*ipc.ChainSpec_GenesisAccount, 1)
	accounts[0] = &ipc.ChainSpec_GenesisAccount{}
	accounts[0].PublicKey, err = base64.StdEncoding.DecodeString(
		"s8qP7TauBe0WoHUDEKyFR99XM6q7aGzacLa6M6vHtO0=")
	accounts[0].Balance = &state.BigInt{Value: "50000000000", BitWidth: 512}
	accounts[0].BondedAmount = &state.BigInt{Value: "1000000", BitWidth: 512}
	expected.Accounts = accounts

	// CostTable

	expected.Costs = &ipc.ChainSpec_CostTable{}
	wasmTable := ipc.ChainSpec_CostTable_WasmCosts{
		Regular:        1,
		Div:            16,
		Mul:            4,
		Mem:            2,
		InitialMem:     4096,
		GrowMem:        8192,
		Memcpy:         1,
		MaxStackHeight: 65536,
		OpcodesMul:     3,
		OpcodesDiv:     8,
	}
	expected.Costs.Wasm = &wasmTable

	return &expected, nil
}

func TestReadGenesisConfig(t *testing.T) {
	got, err := ReadGenesisConfig("../resources/manifest.toml")
	require.Nil(t, err)
	require.NotNil(t, got)
	expected, err := genesisConfigMock()
	require.Nil(t, err)

	require.Equal(t, expected.Name, got.Name)
	require.Equal(t, expected.Timestamp, got.Timestamp)
	if !reflect.DeepEqual(expected.ProtocolVersion, got.ProtocolVersion) {
		t.Errorf("Protocol versions differ. expected %v, got %v",
			expected.ProtocolVersion, got.ProtocolVersion)
	}
	require.Equal(t, expected.MintInstaller, got.MintInstaller)
	require.Equal(t, expected.PosInstaller, got.PosInstaller)

	if !reflect.DeepEqual(expected.Accounts, got.Accounts) {
		t.Errorf("Genesis Accounts differ. expected %v, got %v",
			expected.Accounts, got.Accounts)
	}
	if !reflect.DeepEqual(expected.Costs, got.Costs) {
		t.Errorf("Costs table differ. expected %v, got %v",
			expected.Costs, got.Costs)
	}
}
