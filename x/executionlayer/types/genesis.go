package types

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
)

// GenesisState : the executionlayer state that must be provided at genesis.
type GenesisState struct {
	GenesisConf GenesisConf `json:"genesis_conf"`
}

// GenesisConf : the executionlayer configuration that must be provided at genesis.
type GenesisConf struct {
	Genesis   Genesis   `json:"genesis"`
	WasmCosts WasmCosts `json:"wasm_costs"`
}

// Genesis : Chain Genesis information
type Genesis struct {
	Name            string    `json:"name"`
	Timestamp       uint64    `json:"timestamp"`
	MintWasm        []byte    `json:"mint_wasm"`
	PosWasm         []byte    `json:"pos_wasm"`
	Accounts        []Account `json:"accounts"`
	ProtocolVersion string    `json:"protocol_version"`
}

// Account : Genesis Account Information.
type Account struct {
	// PublicKey : base64 encoded public key string
	PublicKey           PublicKey `json:"public_key"`
	InitialBalance      string    `json:"initial_balance"`
	InitialBondedAmount string    `json:"initial_bonded_amount"`
}

// WasmCosts : CasperLabs EE Wasm Cost table
type WasmCosts struct {
	Regular           uint32 `json:"regular"`
	DivMultiplier     uint32 `json:"div_multiplier"`
	MulMultiplier     uint32 `json:"mul_multiplier"`
	MemMultiplier     uint32 `json:"mem_multiplier"`
	MemInitialPages   uint32 `json:"mem_initial_pages"`
	MemGrowPerPage    uint32 `json:"mem_grow_per_page"`
	MemCopyPerByte    uint32 `json:"mem_copy_per_byte"`
	MaxStackHeight    uint32 `json:"max_stack_height"`
	OpcodesMultiplier uint32 `json:"opcodes_multiplier"`
	OpcodesDivisor    uint32 `json:"opcodes_divisor"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(genesisConf GenesisConf) GenesisState {
	return GenesisState{GenesisConf: genesisConf}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	mintWasm, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.nodef/contracts/mint_install.wasm"))
	if err != nil {
		panic(err)
	}
	posWasm, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.nodef/contracts/pos_install.wasm"))
	if err != nil {
		panic(err)
	}
	genesisConf := GenesisConf{
		Genesis: Genesis{
			Name:            "friday-devnet",
			Timestamp:       0,
			MintWasm:        mintWasm,
			PosWasm:         posWasm,
			Accounts:        make([]Account, 0),
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
	}
	return NewGenesisState(genesisConf)
}

// ValidateGenesis :
func ValidateGenesis(data GenesisState) error {
	_, err := ToChainSpecGenesisConfig(data.GenesisConf)
	return err
}

func ToChainSpecGenesisConfig(config GenesisConf) (*ipc.ChainSpec_GenesisConfig, error) {
	pv, err := toProtocolVersion(config.Genesis.ProtocolVersion)
	if err != nil {
		return nil, err
	}

	var accounts []*ipc.ChainSpec_GenesisAccount
	if n := len(config.Genesis.Accounts); n != 0 {
		accounts = make([]*ipc.ChainSpec_GenesisAccount, n)
		for i, v := range config.Genesis.Accounts {
			account := toChainSpecGenesisAccount(v)
			accounts[i] = &account
		}
	}

	chainSpecConfig := ipc.ChainSpec_GenesisConfig{
		Name:            config.Genesis.Name,
		Timestamp:       config.Genesis.Timestamp,
		ProtocolVersion: pv,
		MintInstaller:   config.Genesis.MintWasm,
		PosInstaller:    config.Genesis.PosWasm,
		Accounts:        accounts,
		Costs:           toCostTable(config.WasmCosts),
	}
	return &chainSpecConfig, nil
}

func toProtocolVersion(pvString string) (*state.ProtocolVersion, error) {
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
	genesisAccount.PublicKey = account.PublicKey
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

// value validation is performed in ExecutionEngine
func toBigInt(value string) state.BigInt {
	ret := state.BigInt{}
	ret.Value = value
	ret.BitWidth = 512
	return ret
}
