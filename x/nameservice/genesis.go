package nameservice

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type GenesisStateStorage struct {
	UnitAccountArr []MsgSetAccount `json:"accountarr"`
}

type GenesisStateLoad struct {
	UnitAccountArr []UnitAccount `json:"accountarr"`
}

func NewGenesisState(accountRec []UnitAccount) GenesisStateLoad {
	return GenesisStateLoad{UnitAccountArr: nil}
}

func ValidateGenesis(data GenesisStateStorage) error {
	for _, record := range data.UnitAccountArr {
		if record.ID == "" {
			return fmt.Errorf("Invalid UnitAccount: ID: %s. Error: Missing id", record.ID)
		}
		if record.Address.String() == "" {
			return fmt.Errorf("Invalid UnitAccount: Address: %s. Error: Missing Address", record.Address.String())
		}
	}
	return nil
}

func DefaultGenesisState() GenesisStateLoad {
	return GenesisStateLoad{
		UnitAccountArr: []UnitAccount{},
	}
}

func InitGenesis(ctx sdk.Context, k AccountKeeper, data GenesisStateStorage) []abci.ValidatorUpdate {
	for _, record := range data.UnitAccountArr {
		k.SetUnitAccount(ctx, record.ID, record.Address)
	}
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k AccountKeeper) GenesisStateStorage {
	var records []MsgSetAccount
	iterator := k.GetAccountIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		name := string(iterator.Key())
		var acc UnitAccount
		acc = k.GetUnitAccount(ctx, name)

		strname, _ := acc.ID.ToString()
		convertedAcc := MsgSetAccount{
			ID:      strname,
			Address: acc.Address,
		}
		records = append(records, convertedAcc)
	}
	return GenesisStateStorage{UnitAccountArr: records}
}
