package genaccounts

import (
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	authexported "github.com/hdac-io/friday/x/auth/exported"
	"github.com/hdac-io/friday/x/genaccounts/internal/types"
)

// InitGenesis initializes accounts and deliver genesis transactions
func InitGenesis(ctx sdk.Context, _ *codec.Codec, accountKeeper types.AccountKeeper, genesisState GenesisState) {
	genesisState.Sanitize()

	// load the accounts
	for _, gacc := range genesisState {
		acc := gacc.ToAccount()
		acc = accountKeeper.NewAccount(ctx, acc) // set account number
		accountKeeper.SetAccount(ctx, acc)
	}
}

// ExportGenesis exports genesis for all accounts
func ExportGenesis(ctx sdk.Context, accountKeeper types.AccountKeeper) GenesisState {

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	accountKeeper.IterateAccounts(ctx,
		func(acc authexported.Account) (stop bool) {
			account, err := NewGenesisAccountI(acc)
			if err != nil {
				panic(err)
			}
			accounts = append(accounts, account)
			return false
		},
	)

	return accounts
}
