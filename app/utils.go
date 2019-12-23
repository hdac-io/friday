//nolint
package app

import (
	"io"

	"github.com/hdac-io/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/hdac-io/friday/baseapp"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/staking"
)

var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// DONTCOVER

// NewFridayAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewFridayAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (fapp *FridayApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	fapp = NewFridayApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return fapp, fapp.keys[baseapp.MainStoreKey], fapp.keys[staking.StoreKey], fapp.stakingKeeper
}
