package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tm-db"

	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/simapp"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestFridaydExport(t *testing.T) {
	db := db.NewMemDB()
	fapp := NewFridayApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	setGenesis(fapp)

	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewFridayApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	_, _, err := newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// ensure that black listed addresses are properly set in bank keeper
func TestBlackListedAddrs(t *testing.T) {
	db := db.NewMemDB()
	app := NewFridayApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)

	for acc := range maccPerms {
		require.True(t, app.bankKeeper.BlacklistedAddr(app.supplyKeeper.GetModuleAddress(acc)))
	}
}

func setGenesis(fapp *FridayApp) error {

	genesisState := simapp.NewDefaultGenesisState()
	stateBytes, err := codec.MarshalJSONIndent(fapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	fapp.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	fapp.Commit()
	return nil
}
