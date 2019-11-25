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

const (
	resourceDir         = "../resources"
	mintInstallWasmName = "mint_install.wasm"
	posInstallWasmName  = "pos_install.wasm"
	chainSpecfileName   = "manifest.toml"
	genAccountsfileName = "accounts.csv"

	genAccountPublicKey    = "s8qP7TauBe0WoHUDEKyFR99XM6q7aGzacLa6M6vHtO0="
	genAccountBalance      = "50000000000"
	genAccountBondedAmount = "1000000"
)

func genesisConfigMock() (*ipc.ChainSpec_GenesisConfig, error) {
	genesisConfig := ipc.ChainSpec_GenesisConfig{}
	genesisConfig.Name = "friday-devnet"
	genesisConfig.Timestamp = 0
	genesisConfig.ProtocolVersion = &state.ProtocolVersion{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}

	// load mint_install.wasm, pos_install.wasm
	var err error
	genesisConfig.MintInstaller, err = ioutil.ReadFile(resourceDir + "/" + mintInstallWasmName)
	if err != nil {
		return nil, err
	}
	genesisConfig.PosInstaller, err = ioutil.ReadFile(resourceDir + "/" + posInstallWasmName)
	if err != nil {
		return nil, err
	}

	// GenesisAccount
	accounts := make([]*ipc.ChainSpec_GenesisAccount, 1)
	accounts[0] = &ipc.ChainSpec_GenesisAccount{}
	accounts[0].PublicKey, err = base64.StdEncoding.DecodeString(genAccountPublicKey)
	accounts[0].Balance = &state.BigInt{Value: genAccountBalance, BitWidth: 512}
	accounts[0].BondedAmount = &state.BigInt{Value: genAccountBondedAmount, BitWidth: 512}
	genesisConfig.Accounts = accounts

	// CostTable
	genesisConfig.Costs = &ipc.ChainSpec_CostTable{}
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
	genesisConfig.Costs.Wasm = &wasmTable

	return &genesisConfig, nil
}

func TestReadChainSpec(t *testing.T) {
	// invalid path
	got, err := readChainSpec("")
	require.NotNil(t, err)
	require.Nil(t, got)

	// valid path
	got, err = readChainSpec(resourceDir + "/" + chainSpecfileName)
	require.Nil(t, err)
	require.NotNil(t, got)

	expected := GenesisConf{
		Genesis: Genesis{
			Name:                "friday-devnet",
			Timestamp:           0,
			MintCodePath:        mintInstallWasmName,
			PosCodePath:         posInstallWasmName,
			InitialAccountsPath: genAccountsfileName,
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
	// invalid path
	got, err := readGenesisAccountsCsv("")
	require.NotNil(t, err)
	require.Empty(t, got)

	// valid path
	got, err = readGenesisAccountsCsv(resourceDir + "/" + genAccountsfileName)
	require.Nil(t, err)
	require.Equal(t, 1, len(got))

	expected := Account{
		publicKey:           genAccountPublicKey,
		initialBalance:      genAccountBalance,
		initialBondedAmount: genAccountBondedAmount,
	}
	if !reflect.DeepEqual(expected, got[0]) {
		t.Errorf("Bad accounts.csv, expected %v, got %v", expected, got)
	}
}

func TestFromAccount(t *testing.T) {
	// mock account
	account := Account{
		publicKey:           genAccountPublicKey,
		initialBalance:      genAccountBalance,
		initialBondedAmount: genAccountBondedAmount,
	}

	got, err := fromAccount(account)
	require.Nil(t, err)
	require.NotNil(t, got)

	expected := ipc.ChainSpec_GenesisAccount{}
	expected.PublicKey, err = base64.StdEncoding.DecodeString(genAccountPublicKey)
	require.Nil(t, err)
	expected.Balance = &state.BigInt{Value: genAccountBalance, BitWidth: 512}
	expected.BondedAmount = &state.BigInt{Value: genAccountBondedAmount, BitWidth: 512}

	if !reflect.DeepEqual(expected, *got) {
		t.Errorf("Bad Account parsing. expected %v, got %v", expected, got)
	}
}

func TestParseProtocolVersion(t *testing.T) {
	// empty string
	got, err := parseProtocolVersion("")
	require.NotNil(t, err)
	require.Nil(t, got)

	// just a number
	got, err = parseProtocolVersion("123")
	require.NotNil(t, err)
	require.Nil(t, got)

	// trailing dot
	got, err = parseProtocolVersion("1.0.0.")
	require.NotNil(t, err)
	require.Nil(t, got)

	// too many digit
	got, err = parseProtocolVersion("1.0.0.0")
	require.NotNil(t, err)
	require.Nil(t, got)

	// valid case
	got, err = parseProtocolVersion("123.456.789")
	require.Nil(t, err)
	expected := state.ProtocolVersion{Major: 123, Minor: 456, Patch: 789}
	if !reflect.DeepEqual(expected, *got) {
		t.Errorf("Bad protocol version parsing. expected %v, got %v", expected, got)
	}
}

func TestReadGenesisConfig(t *testing.T) {
	// invalid path
	got, err := ReadGenesisConfig("")
	require.NotNil(t, err)
	require.Nil(t, got)

	// valid path
	got, err = ReadGenesisConfig(resourceDir + "/" + chainSpecfileName)
	require.Nil(t, err)
	require.NotNil(t, got)
	expected, err := genesisConfigMock()
	require.Nil(t, err)

	// validation
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
