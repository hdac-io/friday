package types

import (
	"os"
	"strconv"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/util"

	sdk "github.com/hdac-io/friday/types"
)

// GenesisState : the executionlayer state that must be provided at genesis.
type GenesisState struct {
	GenesisConf GenesisConf `json:"genesis_conf"`
	Accounts    []Account   `json:"accounts"`
	ChainName   string      `json:"chain_name"`
}

// GenesisConf : the executionlayer configuration that must be provided at genesis.
type GenesisConf struct {
	Genesis      Genesis      `json:"genesis"`
	WasmCosts    WasmCosts    `json:"wasm_costs"`
	DeployConfig DeployConfig `json:"deploy_config"`
}

// Genesis : Chain Genesis information
type Genesis struct {
	Timestamp       uint64 `json:"timestamp"`
	MintWasm        []byte `json:"mint_wasm"`
	PosWasm         []byte `json:"pos_wasm"`
	ProtocolVersion string `json:"protocol_version"`
}

// Account : Genesis Account Information.
type Account struct {
	Address             sdk.AccAddress `json:"address"`
	InitialBalance      string         `json:"initial_balance"`
	InitialBondedAmount string         `json:"initial_bonded_amount"`
}

// WasmCosts : CasperLabs EE Wasm Cost table
type WasmCosts struct {
	Regular           uint32 `json:"regular" toml:"regular"`
	DivMultiplier     uint32 `json:"div_multiplier" toml:"div-multiplier"`
	MulMultiplier     uint32 `json:"mul_multiplier" toml:"mul-multiplier"`
	MemMultiplier     uint32 `json:"mem_multiplier" toml:"mem-multiplier"`
	MemInitialPages   uint32 `json:"mem_initial_pages" toml:"mem-initial-pages"`
	MemGrowPerPage    uint32 `json:"mem_grow_per_page" toml:"mem-grow-per-page"`
	MemCopyPerByte    uint32 `json:"mem_copy_per_byte" toml:"mem-copy-per-byte"`
	MaxStackHeight    uint32 `json:"max_stack_height" toml:"max-stack-height"`
	OpcodesMultiplier uint32 `json:"opcodes_multiplier" toml:"opcodes-multiplier"`
	OpcodesDivisor    uint32 `json:"opcodes_divisor" toml:"opcodes-divisor"`
}

type DeployConfig struct {
	MaxTtlMillis    uint32 `json:"max-ttl-millis" toml:"max-ttl-millis"`
	MaxDependencies uint32 `json:"max-dependencies" toml:"max-dependencies"`
}

const (
	mintCodePath = "$HOME/.nodef/contracts/hdac_mint_install.wasm"
	posCodePath  = "$HOME/.nodef/contracts/pop_install.wasm"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(genesisConf GenesisConf, accounts []Account, chainName string) GenesisState {
	return GenesisState{GenesisConf: genesisConf, Accounts: accounts, ChainName: chainName}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	genesisConf := GenesisConf{
		Genesis: Genesis{
			Timestamp:       0,
			MintWasm:        util.LoadWasmFile(os.ExpandEnv(mintCodePath)),
			PosWasm:         util.LoadWasmFile(os.ExpandEnv(posCodePath)),
			ProtocolVersion: "1.0.0",
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
		DeployConfig: DeployConfig{
			MaxTtlMillis:    86400000,
			MaxDependencies: 10,
		},
	}
	return NewGenesisState(genesisConf, nil, "friday-devnet")
}

// ValidateGenesis :
func ValidateGenesis(data GenesisState) error {
	_, err := ToChainSpecGenesisConfig(data)
	return err
}

func ToChainSpecGenesisConfig(gs GenesisState) (*ipc.ChainSpec_GenesisConfig, error) {
	config := gs.GenesisConf
	pv, err := ToProtocolVersion(config.Genesis.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	var accounts []*ipc.ChainSpec_GenesisAccount
	if n := len(gs.Accounts); n != 0 {
		accounts = make([]*ipc.ChainSpec_GenesisAccount, n)
		for i, v := range gs.Accounts {
			account := toChainSpecGenesisAccount(v)
			accounts[i] = &account
		}
	}

	chainSpecConfig := ipc.ChainSpec_GenesisConfig{
		Name:            gs.ChainName,
		Timestamp:       config.Genesis.Timestamp,
		ProtocolVersion: pv,
		MintInstaller:   config.Genesis.MintWasm,
		PosInstaller:    config.Genesis.PosWasm,
		Accounts:        accounts,
		Costs:           toCostTable(config.WasmCosts),
		DeployConfig:    toDeployConfig(config.DeployConfig),
	}

	return &chainSpecConfig, nil
}

func ToProtocolVersion(pvString string) (*state.ProtocolVersion, error) {
	splittedProtocolVer := strings.Split(pvString, ".")
	if len(splittedProtocolVer) != 3 {
		return nil, ErrProtocolVersionParse(DefaultCodespace, pvString)
	}
	major, err := strconv.ParseUint(splittedProtocolVer[0], 10, 32)
	if err != nil {
		return nil, ErrProtocolVersionParse(DefaultCodespace, pvString)
	}
	minor, err := strconv.ParseUint(splittedProtocolVer[1], 10, 32)
	if err != nil {
		return nil, ErrProtocolVersionParse(DefaultCodespace, pvString)
	}
	patch, err := strconv.ParseUint(splittedProtocolVer[2], 10, 32)
	if err != nil {
		return nil, ErrProtocolVersionParse(DefaultCodespace, pvString)
	}

	return &state.ProtocolVersion{
			Major: uint32(major), Minor: uint32(minor), Patch: uint32(patch)},
		nil
}

func toChainSpecGenesisAccount(account Account) ipc.ChainSpec_GenesisAccount {
	balance := toBigInt(account.InitialBalance)
	bondedAmount := toBigInt(account.InitialBondedAmount)

	genesisAccount := ipc.ChainSpec_GenesisAccount{}

	genesisAccount.PublicKey = account.Address.ToEEAddress().Bytes()
	genesisAccount.Balance = &balance
	genesisAccount.BondedAmount = &bondedAmount

	return genesisAccount
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

func toDeployConfig(deployConfig DeployConfig) *ipc.ChainSpec_DeployConfig {
	return &ipc.ChainSpec_DeployConfig{
		MaxTtlMillis:    deployConfig.MaxTtlMillis,
		MaxDependencies: deployConfig.MaxDependencies,
	}
}

// value validation is performed in ExecutionEngine
func toBigInt(value string) state.BigInt {
	ret := state.BigInt{}
	ret.Value = value
	ret.BitWidth = 512
	return ret
}
