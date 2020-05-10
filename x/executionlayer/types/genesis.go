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
	Validators  []Validator `json:"validators"`
	StateInfos  []string    `json:"state_infos"`
}

// GenesisConf : the executionlayer configuration that must be provided at genesis.
type GenesisConf struct {
	Genesis       Genesis       `json:"genesis"`
	WasmCosts     WasmCosts     `json:"wasm_costs"`
	DeployConfig  DeployConfig  `json:"deploy_config"`
	HighwayConfig HighwayConfig `json:"highway_config"`
}

// Genesis : Chain Genesis information
type Genesis struct {
	Timestamp           uint64 `json:"timestamp"`
	MintWasm            []byte `json:"mint_wasm"`
	PosWasm             []byte `json:"pos_wasm"`
	StandardPaymentWasm []byte `json:"standrad_payment_wasm`
	ProtocolVersion     string `json:"protocol_version"`
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
	MaxTtlMillis      uint32 `json:"max-ttl-millis" toml:"max-ttl-millis"`
	MaxDependencies   uint32 `json:"max-dependencies" toml:"max-dependencies"`
	MaxBlockSizeBytes uint32 `json:"max-block-size-bytes" toml:"max-block-size-bytes"`
	MaxBlockCost      uint64 `json:"max-block-cost" toml:"max-block-cost"`
}

type HighwayConfig struct {
	GenesisEraStartTimestamp   uint64 `json:"genesis-era-start" toml:"genesis-era-start"`
	EraDurationMillis          uint64 `json:"era-duration" toml:"era-duration"`
	BookingDurationMillis      uint64 `json:"booking-duration" toml:"booking-duration"`
	EntropyDurationMillis      uint64 `json:"entropy-duration" toml:"entropy-duration"`
	VotingPeriodDurationMillis uint64 `json:"voting-period-duration" toml:"voting-period-duration"`
	VotingPeriodSummitLevel    uint32 `json:"voting-period-summit-level" toml:"voting-period-summit-level"`
	Ftt                        uint32 `json:"ftt" toml:"ftt"`
}

const (
	mintCodePath            = "$HOME/.nodef/contracts/hdac_mint_install.wasm"
	posCodePath             = "$HOME/.nodef/contracts/pop_install.wasm"
	standardPaymentCodePath = "$HOME/.nodef/contracts/standard_payment_install.wasm"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(genesisConf GenesisConf, accounts []Account, chainName string, validators Validators, stateInfos []string) GenesisState {
	return GenesisState{GenesisConf: genesisConf, Accounts: accounts, ChainName: chainName, Validators: validators, StateInfos: stateInfos}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	genesisConf := GenesisConf{
		Genesis: Genesis{
			Timestamp:           0,
			MintWasm:            util.LoadWasmFile(os.ExpandEnv(mintCodePath)),
			PosWasm:             util.LoadWasmFile(os.ExpandEnv(posCodePath)),
			StandardPaymentWasm: util.LoadWasmFile(os.ExpandEnv(standardPaymentCodePath)),
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
		DeployConfig: DeployConfig{
			MaxTtlMillis:      86400000,
			MaxDependencies:   10,
			MaxBlockSizeBytes: 10485760,
			MaxBlockCost:      0,
		},
		HighwayConfig: HighwayConfig{
			GenesisEraStartTimestamp:   1583712000000,
			EraDurationMillis:          604800000,
			BookingDurationMillis:      864000000,
			EntropyDurationMillis:      10800000,
			VotingPeriodDurationMillis: 172800000,
			VotingPeriodSummitLevel:    0,
			Ftt:                        0,
		},
	}
	return NewGenesisState(genesisConf, nil, "friday-devnet", nil, nil)
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
		HighwayConfig:   toHighwayConfig(config.HighwayConfig),
		StateInfos:      gs.StateInfos,
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

	genesisAccount.PublicKey = account.Address.Bytes()
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
		MaxTtlMillis:      deployConfig.MaxTtlMillis,
		MaxDependencies:   deployConfig.MaxDependencies,
		MaxBlockSizeBytes: deployConfig.MaxBlockSizeBytes,
		MaxBlockCost:      deployConfig.MaxBlockCost,
	}
}

func toHighwayConfig(highwayConfig HighwayConfig) *ipc.ChainSpec_HighwayConfig {
	return &ipc.ChainSpec_HighwayConfig{
		GenesisEraStartTimestamp:   highwayConfig.GenesisEraStartTimestamp,
		EraDurationMillis:          highwayConfig.EraDurationMillis,
		BookingDurationMillis:      highwayConfig.BookingDurationMillis,
		EntropyDurationMillis:      highwayConfig.EntropyDurationMillis,
		VotingPeriodDurationMillis: highwayConfig.VotingPeriodDurationMillis,
		VotingPeriodSummitLevel:    highwayConfig.VotingPeriodSummitLevel,
		Ftt:                        float64(highwayConfig.Ftt),
	}
}

// value validation is performed in ExecutionEngine
func toBigInt(value string) state.BigInt {
	ret := state.BigInt{}
	ret.Value = value
	ret.BitWidth = 512
	return ret
}
