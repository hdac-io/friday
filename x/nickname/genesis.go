package nickname

import (
	"fmt"

	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"
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
		if record.Nickname.Equal(NewName("")) {
			return fmt.Errorf("Invalid UnitAccount!\nName: %s. Error: Missing id", record.Nickname.MustToString())
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

func InitGenesis(ctx sdk.Context, k NicknameKeeper, data GenesisStateStorage) []abci.ValidatorUpdate {
	for _, record := range data.UnitAccountArr {
		k.SetNickname(ctx, record.Nickname.MustToString(), record.Address)
	}
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k NicknameKeeper) GenesisStateStorage {
	var records []MsgSetAccount
	iterator := k.GetAccountIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		name := string(iterator.Key())
		var acc UnitAccount
		acc = k.GetUnitAccount(ctx, name)

		convertedAcc := NewMsgSetAccount(acc.Nickname, acc.Address)
		records = append(records, convertedAcc)
	}
	return GenesisStateStorage{UnitAccountArr: records}
}
