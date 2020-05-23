package types

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
)

// GenesisState - crisis genesis state
type GenesisState struct {
	ConstantFee sdk.Coin `json:"constant_fee" yaml:"constant_fee"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(constantFee sdk.Coin) GenesisState {
	return GenesisState{
		ConstantFee: constantFee,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// ValidateGenesis - validate crisis genesis data
func ValidateGenesis(data GenesisState) error {
	if !data.ConstantFee.IsPositive() {
		return fmt.Errorf("constant fee must be positive: %s", data.ConstantFee)
	}
	return nil
}
