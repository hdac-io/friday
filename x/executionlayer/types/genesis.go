package types

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	toml "github.com/pelletier/go-toml"
)

// Genesis : Chain Genesis information
type Genesis struct {
	Name                string `toml:"name"`
	Timestamp           uint64 `toml:"timestamp"`
	MintCodePath        string `toml:"mint-code-path"`
	PosCodePath         string `toml:"pos-code-path"`
	InitialAccountsPath string `toml:"initial-accounts-path"`
	ProtocolVersion     string `toml:"protocol-version"`
}

// WasmCosts : CasperLabs EE Wasm Cost table
type WasmCosts struct {
	Regular           uint32 `toml:"regular"`
	DivMultiplier     uint32 `toml:"div-multiplier"`
	MulMultiplier     uint32 `toml:"mul-multiplier"`
	MemMultiplier     uint32 `toml:"mem-multiplier"`
	MemInitialPages   uint32 `toml:"mem-initial-pages"`
	MemGrowPerPage    uint32 `toml:"mem-grow-per-page"`
	MemCopyPerByte    uint32 `toml:"mem-copy-per-byte"`
	MaxStackHeight    uint32 `toml:"max-stack-height"`
	OpcodesMultiplier uint32 `toml:"opcodes-multiplier"`
	OpcodesDivisor    uint32 `toml:"opcodes-divisor"`
}

// GenesisConf :
type GenesisConf struct {
	Genesis   Genesis   `toml:"genesis"`
	WasmCosts WasmCosts `toml:"wasm-costs"`
}

// Account : Genesis Account Information.
type Account struct {
	publicKey           string
	initialBalance      string
	initialBondedAmount string
}

// ReadChainSpec : Load Chain Specification from the toml file.
func readChainSpec(chainSpecPath string) (*GenesisConf, error) {
	if _, err := os.Stat(chainSpecPath); os.IsNotExist(err) {
		fmt.Fprintf(
			os.Stderr, "ReadChainSpec: \"%s\" does not exist\n", chainSpecPath)
		return nil, err
	}

	tree, err := toml.LoadFile(chainSpecPath)
	if err != nil {
		return nil, err
	}

	genesisConf := GenesisConf{}
	if tree.Unmarshal(&genesisConf); err != nil {
		return nil, err
	}

	return &genesisConf, nil
}

// ReadGenesisAccountsCsv : parse accounts.csv corresponding path and
// load into Account array
func readGenesisAccountsCsv(accountsCsvPath string) ([]Account, error) {
	content, err := ioutil.ReadFile(accountsCsvPath)
	if err != nil {
		return nil, err
	}

	splittedContent := strings.Split(string(content), ",")
	splittedContentLen := len(splittedContent)

	if splittedContentLen%3 != 0 {
		return nil, ErrMalforemdAccountsCsv(DefaultCodespace)
	}

	accounts := make([]Account, splittedContentLen/3)
	for i := 0; i < splittedContentLen; i += 3 {
		accounts[i] = Account{
			publicKey:           splittedContent[i],
			initialBalance:      splittedContent[i+1],
			initialBondedAmount: splittedContent[i+2],
		}
	}

	return accounts, err
}

func toGenesisAccount(account Account) (*ipc.ChainSpec_GenesisAccount, error) {
	publicKey, err := base64.StdEncoding.DecodeString(account.publicKey)
	if err != nil {
		return nil, err
	}
	// TODO : value vaiadation, define error code
	balance := state.BigInt{
		Value:    account.initialBalance,
		BitWidth: 512,
	}
	bondedAmount := state.BigInt{
		Value:    account.initialBondedAmount,
		BitWidth: 512,
	}

	return &ipc.ChainSpec_GenesisAccount{
		PublicKey:    publicKey,
		Balance:      &balance,
		BondedAmount: &bondedAmount,
	}, nil
}

func toCostTable(wasmCosts WasmCosts) *ipc.ChainSpec_CostTable {
	costTable := ipc.ChainSpec_CostTable{}
	costTable.Wasm = &ipc.ChainSpec_CostTable_WasmCosts{}
	costTable.Wasm.Regular = wasmCosts.Regular
	costTable.Wasm.Div = wasmCosts.DivMultiplier
	costTable.Wasm.Mul = wasmCosts.MulMultiplier
	costTable.Wasm.Mem = wasmCosts.MemMultiplier
	costTable.Wasm.InitialMem = wasmCosts.MemInitialPages
	costTable.Wasm.GrowMem = wasmCosts.MemGrowPerPage
	costTable.Wasm.Memcpy = wasmCosts.MemCopyPerByte
	costTable.Wasm.MaxStackHeight = wasmCosts.MaxStackHeight
	costTable.Wasm.OpcodesMul = wasmCosts.OpcodesMultiplier
	costTable.Wasm.OpcodesDiv = wasmCosts.OpcodesDivisor
	return &costTable
}

func toProtocolVersion(pvString string) (*state.ProtocolVersion, error) {
	splittedProtocolVer := strings.Split(pvString, ".")
	if len(splittedProtocolVer) != 3 {
		return nil, ErrProtocolVersionParse(DefaultCodespace)
	}
	major, err := strconv.ParseUint(splittedProtocolVer[0], 10, 32)
	if err != nil {
		return nil, ErrProtocolVersionParse(DefaultCodespace)
	}
	minor, err := strconv.ParseUint(splittedProtocolVer[1], 10, 32)
	if err != nil {
		return nil, ErrProtocolVersionParse(DefaultCodespace)
	}
	patch, err := strconv.ParseUint(splittedProtocolVer[2], 10, 32)
	if err != nil {
		return nil, ErrProtocolVersionParse(DefaultCodespace)
	}

	return &state.ProtocolVersion{
			Major: uint32(major), Minor: uint32(minor), Patch: uint32(patch)},
		nil
}

// ReadGenesisConfig :
func ReadGenesisConfig(chainSpecPath string) (*ipc.ChainSpec_GenesisConfig, error) {
	genesisConfig := ipc.ChainSpec_GenesisConfig{}
	chainSpec, err := readChainSpec(chainSpecPath)
	if err != nil {
		return nil, err
	}

	genesisConfig.Name = chainSpec.Genesis.Name
	genesisConfig.Timestamp = chainSpec.Genesis.Timestamp

	if genesisConfig.ProtocolVersion, err = toProtocolVersion(
		chainSpec.Genesis.ProtocolVersion); err != nil {
		return nil, err
	}

	if err = os.Chdir(filepath.Dir(chainSpecPath)); err != nil {
		return nil, err
	}

	if genesisConfig.MintInstaller, err = ioutil.ReadFile(
		chainSpec.Genesis.MintCodePath); err != nil {
		return nil, err
	}
	if genesisConfig.PosInstaller, err = ioutil.ReadFile(
		chainSpec.Genesis.PosCodePath); err != nil {
		return nil, err
	}

	accounts, err := readGenesisAccountsCsv(chainSpec.Genesis.InitialAccountsPath)
	if err != nil {
		return nil, err
	}

	genesisAccounts := make([]*ipc.ChainSpec_GenesisAccount, len(accounts))
	for i, v := range accounts {
		genesisAccount, err := toGenesisAccount(v)
		if err != nil {
			return nil, err
		}
		genesisAccounts[i] = genesisAccount
	}
	genesisConfig.Accounts = genesisAccounts

	genesisConfig.Costs = toCostTable(chainSpec.WasmCosts)

	return &genesisConfig, nil
}
