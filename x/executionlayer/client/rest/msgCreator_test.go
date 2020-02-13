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

	"github.com/hdac-io/friday/x/executionlayer/types"
)

func prepare() (fromAddr, receipAddr string, w http.ResponseWriter, clictx context.CLIContext, basereq rest.BaseReq) {
	fromAddr = "friday15evpva2u57vv6l5czehyk69s0wnq9hrkqulwfz"
	receipAddr = "friday1y2dx0evs5k6hxuhfrfdmm7wcwsrqr073htghpv"

	w = httptest.NewRecorder()
	cdc := codec.New()
	clictx = context.NewCLIContext().WithCodec(cdc)

	basereq = rest.BaseReq{
		From:    fromAddr,
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
		ChainID:                    basereq.ChainID,
		Memo:                       basereq.Memo,
		TokenContractAddress:       fromAddr,
		SenderAddressOrNickname:    fromAddr,
		RecipientAddressOrNickname: receipAddr,
		Amount:                     20_000_000,
		GasPrice:                   gas,
		Fee:                        10_000_000,
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
		ChainID:           basereq.ChainID,
		Memo:              basereq.Memo,
		AddressOrNickname: fromAddr,
		Amount:            100_000_000,
		Fee:               10_000_000,
		GasPrice:          gas,
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
		ChainID:           basereq.ChainID,
		Memo:              basereq.Memo,
		AddressOrNickname: fromAddr,
		Amount:            100_000_000,
		Fee:               10_000_000,
		GasPrice:          gas,
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

func TestRESTGetValidator(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/validator?address=%s", types.ModuleName, fromAddr), nil)
	res, err := getValidatorQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetValidators(t *testing.T) {
	_, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/validator", types.ModuleName), nil)
	res, err := getValidatorQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTCreateValidator(t *testing.T) {
	fromAddr, _, writer, clictx, basereq := prepare()

	gas, _ := strconv.ParseUint(basereq.Gas, 10, 64)
	createValidatorReq := createValidatorReq{
		ChainID:                    basereq.ChainID,
		ValidatorAddressOrNickName: fromAddr,
		ConsPubKey:                 "fridayvalconspub16jrl8jvqq9k957nfd43n2dnyxc6nsazpgf5yuwtzfe6kku63ga6nvtmcdeg92vj4gy4kkd62vd69vvnhx935w5zpw9ex7733tft8we6evemzke66xv4ks56gfdvx66ndfye5x5z9fs6j74z6g3u4zdzd0p8hw6mr24k8wjzx0ghhz5z8vdm92vjs2e8xwdn5xpvxu56fvejnj7t6wsens5gwxlen9",
		Description:                types.NewDescription("moniker", "identity", "https://test.io", "details"),
		Gas:                        gas,
		Memo:                       basereq.Memo,
	}

	body := clictx.Codec.MustMarshalJSON(createValidatorReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/validator", types.ModuleName), bytes.NewReader((body)))

	outputCreateValidatorReq, msgs, err := createValidatorMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputCreateValidatorReq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTEditValidator(t *testing.T) {
	fromAddr, _, writer, clictx, basereq := prepare()

	gas, _ := strconv.ParseUint(basereq.Gas, 10, 64)
	editValidatorReq := editValidatorReq{
		ChainID:                    basereq.ChainID,
		ValidatorAddressOrNickName: fromAddr,
		Description:                types.NewDescription("moniker", "identity", "https://test.io", "details"),
		Gas:                        gas,
		Memo:                       basereq.Memo,
	}

	body := clictx.Codec.MustMarshalJSON(editValidatorReq)
	req := mustNewRequest(t, "PUT", fmt.Sprintf("/%s/validator", types.ModuleName), bytes.NewReader((body)))

	outputEditValidatorReq, msgs, err := editValidatorMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputEditValidatorReq, basereq)
	require.NotNil(t, msgs)
}

func mustNewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	err = req.ParseForm()
	require.NoError(t, err)
	return req
}
