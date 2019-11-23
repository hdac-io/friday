package types

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml"
)

// Genesis : Chain Genesis information
type Genesis struct {
	Name                string `toml:"name"`
	Timestamp           int64  `toml:"timestamp"`
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

// ReadChainSpec : Load Chain Specification from the toml file.
func ReadChainSpec(chainSpecPath string) (*GenesisConf, error) {
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
