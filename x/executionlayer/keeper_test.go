package executionlayer

import (
	"fmt"
	"os"
	"testing"

	sdk "github.com/hdac-io/friday/types"

	"github.com/stretchr/testify/assert"
)

//-------------------------------------------

func TestGetQueryResult(t *testing.T) {
	input := setupTestInput()
	path := "counter/count"

	keeper := NewExecutionLayerKeeper(input.cdc, sdk.NewKVStoreKey(HashMapStoreKey), sdk.NewKVStoreKey(DeployStoreKey),
		os.ExpandEnv("$HOME/.casperlabs/.casper-node.sock"), "1.0.0")

	parentHash := genesis(keeper)
	parentHash = counterDefine(keeper, parentHash)
	parentHash = counterCall(keeper, parentHash)

	res, err := keeper.GetQueryResult(
		input.ctx,
		parentHash,
		"address", input.genesisAddress, path)

	if err != nil {
		fmt.Println(err.Error())
		panic("Fail to execute")
	}

	fmt.Println(res)

	assert.NotNil(t, res)
}
