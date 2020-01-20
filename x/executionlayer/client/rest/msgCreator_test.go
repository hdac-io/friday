package rest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	"github.com/hdac-io/friday/types/rest"
	"github.com/stretchr/testify/require"

	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/x/executionlayer/types"
)

func prepare() (fromAddr, receipAddr string, w http.ResponseWriter, clictx context.CLIContext, basereq rest.BaseReq) {
	fromAddr = "02014a87d1ec490005f85bb4296596ed741411f673a35317543439971c7c7731bb"
	receipAddr = "0216cde7d343c4bb6f8236c85b6f6a541e2d13ce8a306664714384bf53fd9d14e5"

	w = httptest.NewRecorder()
	cdc := codec.New()
	clictx = context.NewCLIContext().WithCodec(cdc)

	basereq = rest.BaseReq{
		From:    sdk.AccAddress(sdk.MustGetSecp256k1FromRawHexString(fromAddr).Address()).String(),
		ChainID: "monday-0001",
		Gas:     fmt.Sprint(60_000_000),
		Memo:    "",
	}

	return
}

func TestRESTTransfer(t *testing.T) {
	fromAddr, receipAddr, writer, clictx, basereq := prepare()

	// Body
	gas, _ := strconv.ParseUint(basereq.Gas, 10, 64)
	transReq := transferReq{
		ChainID:               basereq.ChainID,
		Memo:                  basereq.Memo,
		TokenContractAddress:  fromAddr,
		SenderPubkeyOrName:    fromAddr,
		RecipientPubkeyOrName: receipAddr,
		Amount:                20_000_000,
		GasPrice:              gas,
		Fee:                   10_000_000,
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(transReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/transfer", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := transferMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTBond(t *testing.T) {
	fromAddr, _, writer, clictx, basereq := prepare()

	// Body
	gas, _ := strconv.ParseUint(basereq.Gas, 10, 64)
	bondReq := bondReq{
		ChainID:      basereq.ChainID,
		Memo:         basereq.Memo,
		PubkeyOrName: fromAddr,
		Amount:       100_000_000,
		GasPrice:     gas,
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(bondReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/bond", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := bondUnbondMsgCreator(true, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTUnbond(t *testing.T) {
	fromAddr, _, writer, clictx, basereq := prepare()

	// Body
	gas, _ := strconv.ParseUint(basereq.Gas, 10, 64)
	bondReq := bondReq{
		ChainID:      basereq.ChainID,
		Memo:         basereq.Memo,
		PubkeyOrName: fromAddr,
		Amount:       100_000_000,
		GasPrice:     gas,
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(bondReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/unbond", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := bondUnbondMsgCreator(false, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTBalance(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("/%s/balance?address=%s", types.ModuleName, fromAddr), nil)
	res, err := getBalanceQuerying(writer, clictx, req, types.ModuleName)

	require.NoError(t, err)
	require.NotNil(t, res)
}
func mustNewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	err = req.ParseForm()
	require.NoError(t, err)
	return req
}
