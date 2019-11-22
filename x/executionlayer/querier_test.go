package executionlayer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/hdac-io/friday/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func setup() (testInput, ExecutionLayerKeeper, func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error)) {
	input := setupTestInput()
	keeper := NewExecutionLayerKeeper(input.cdc, sdk.NewKVStoreKey(HashMapStoreKey),
		os.ExpandEnv("$HOME/.casperlabs/.casper-node.sock"), "1.0.0")
	querier := NewQuerier(keeper)

	return input, keeper, querier
}
func TestQueryEEDetail(t *testing.T) {
	input, keeper, querier := setup()
	parentHash := genesis(keeper)
	parentHash = counterDefine(keeper, parentHash)
	parentHash = counterCall(keeper, parentHash)

	queryPath := "counter/count"

	queryData := QueryExecutionLayerDetail{
		StateHash: parentHash,
		KeyType:   "address",
		KeyData:   []byte(input.genesisAddress),
		Path:      queryPath,
	}

	bz, _ := input.cdc.MarshalJSON(queryData)

	query := abci.RequestQuery{
		Path: "querydetail",
		Data: bz,
	}

	handler, err := querier(input.ctx, []string{QueryEEDetail}, query)

	assert.NotNil(t, handler)
	assert.Nil(t, err)
}

func TestQueryBalanceDetail(t *testing.T) {
	input, keeper, querier := setup()
	parentHash := genesis(keeper)

	queryData := QueryGetBalanceDetail{
		StateHash: parentHash,
		Address:   input.genesisAddress,
	}

	bz, _ := input.cdc.MarshalJSON(queryData)

	query := abci.RequestQuery{
		Path: "querybalancedetail",
		Data: bz,
	}

	handler, err := querier(input.ctx, []string{QueryEEBalanceDetail}, query)
	if err != nil {
		panic(err)
	}

	assert.NotNil(t, handler)
	assert.Nil(t, err)
}
