package types

import (
	"encoding/base64"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
)

type GenesisState struct {
	Accounts []Account `json="accounts"`
}

// Account : Genesis Account Information.
type Account struct {
	// PublicKey : base64 encoded public key string
	PublicKey           string `json="public_key"`
	InitialBalance      string `json="initial_balance"`
	InitialBondedAmount string `json="initial_bonded_amount"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(accounts []Account) GenesisState {
	return GenesisState{Accounts: accounts}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(make([]Account, 0))
}

// ValidateGenesis :
func ValidateGenesis(data GenesisState) error {
	// TODO
	return nil
}

func ToGenesisAccount(account Account) (*ipc.ChainSpec_GenesisAccount, error) {
	publicKey, err := base64.StdEncoding.DecodeString(account.PublicKey)
	if err != nil {
		return nil, ErrPublicKeyDecode(DefaultCodespace)
	}
	balance, err := toBigInt(account.InitialBalance)
	if err != nil {
		return nil, err
	}
	bondedAmount, err := toBigInt(account.InitialBondedAmount)
	if err != nil {
		return nil, err
	}

	genesisAccount := ipc.ChainSpec_GenesisAccount{}
	genesisAccount.PublicKey = publicKey
	genesisAccount.Balance = balance
	genesisAccount.BondedAmount = bondedAmount

	return &genesisAccount, nil
}

func toBigInt(value string) (*state.BigInt, error) {
	ret := state.BigInt{}
	// TODO : validation, define bigint conversion err
	ret.Value = value
	ret.BitWidth = 512
	return &ret, nil
}
