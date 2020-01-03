package configuration

import (
	"io/ioutil"
	"path"

	"github.com/hdac-io/friday/x/executionlayer/types"
	toml "github.com/pelletier/go-toml"
)

// ParseGenesisChainSpec loads genesis configuration for CasperLabs execution engine
func ParseGenesisChainSpec(chainSpecPath string) (*types.GenesisConf, error) {
	tree, err := toml.LoadFile(chainSpecPath)
	if err != nil {
		return nil, err
	}
	var genesisConf types.GenesisConf

	// Get genesis
	subTree := tree.Get("genesis").(*toml.Tree)
	if subTree == nil {
		return nil, types.ErrTomlParse(types.DefaultCodespace, "genesis")
	}
	genesis, err := parseGenesisTable(subTree, chainSpecPath)
	if err != nil {
		return nil, err
	}
	genesisConf.Genesis = *genesis

	// Get wasm-costs
	subTree = tree.Get("wasm-costs").(*toml.Tree)
	if subTree == nil {
		return nil, types.ErrTomlParse(types.DefaultCodespace, "wasm-costs")
	}
	err = subTree.Unmarshal(&genesisConf.WasmCosts)
	if err != nil {
		return nil, err
	}

	return &genesisConf, nil
}

func parseGenesisTable(genesisTable *toml.Tree, chainSpecPath string) (*types.Genesis, error) {
	genesis := types.Genesis{}

	if genesisTable.Get("timestamp") == nil {
		return nil, types.ErrTomlParse(types.DefaultCodespace, "timestamp")
	}
	genesis.Timestamp = uint64(genesisTable.Get("timestamp").(int64))

	if genesisTable.Get("protocol-version") == nil {
		return nil, types.ErrTomlParse(types.DefaultCodespace, "protocol-version")
	}
	genesis.ProtocolVersion = genesisTable.Get("protocol-version").(string)

	if genesisTable.Get("mint-code-path") == nil {
		return nil, types.ErrTomlParse(types.DefaultCodespace, "mint-code-path")
	}
	mintCodePath := genesisTable.Get("mint-code-path").(string)
	if !path.IsAbs(mintCodePath) {
		mintCodePath = path.Join(path.Dir(chainSpecPath), mintCodePath)
	}
	mintWasm, err := ioutil.ReadFile(mintCodePath)
	if err != nil {
		return nil, err
	}
	genesis.MintWasm = mintWasm

	if genesisTable.Get("pos-code-path") == nil {
		return nil, types.ErrTomlParse(types.DefaultCodespace, "pos-code-path")
	}
	posCodePath := genesisTable.Get("pos-code-path").(string)
	if !path.IsAbs(posCodePath) {
		posCodePath = path.Join(path.Dir(chainSpecPath), posCodePath)
	}
	posWasm, err := ioutil.ReadFile(posCodePath)
	if err != nil {
		return nil, err
	}
	genesis.PosWasm = posWasm

	return &genesis, nil
}
