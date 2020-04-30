package rest

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/hdac-io/friday/client/context"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	"github.com/hdac-io/friday/types/rest"
	"github.com/stretchr/testify/require"

	"github.com/hdac-io/friday/x/executionlayer/types"

	eeutil "github.com/hdac-io/casperlabs-ee-grpc-go-util/util"
)

var (
	contractPath      = os.ExpandEnv("$HOME/.nodef/contracts")
	counterDefineWasm = "counter_define.wasm"
)

func prepare() (fromAddr, receipAddr string, w http.ResponseWriter, clictx context.CLIContext, basereq rest.BaseReq) {
	fromAddr = "friday1gp2u22697kz6slwa25k2tkhz6st2l0zx3hkfc5wdlpjaauv5czsq2dwu8m"
	receipAddr = "friday16wfryel63g7axeamw68630wglalcnk3llh7z665n05qrrmmfqztqkhgkwv"

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

func TestRESTContractRunWASMFile(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()
	counterBinary := eeutil.LoadWasmFile(path.Join(contractPath, counterDefineWasm))

	contractReq := contractRunReq{
		BaseReq:                       basereq,
		ExecutionType:                 "wasm",
		TokenContractAddressOrKeyName: "",
		Base64EncodedBinary:           base64.StdEncoding.EncodeToString(counterBinary),
		Args:                          "",
		Fee:                           "10000000",
	}

	body := clictx.Codec.MustMarshalJSON(contractReq)
	req := mustNewRequest(t, "POST", "/contract", bytes.NewReader(body))

	outputBasereq, msgs, err := contractRunMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTContractRunContractAddress(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()

	contractReq := contractRunReq{
		BaseReq:                       basereq,
		ExecutionType:                 "uref",
		TokenContractAddressOrKeyName: "fridaycontracturef1v4xev2kdy8hkzvwcadk4a3872lzcyyz8t44du5z2jhz636qduz3sf9mf96",
		Base64EncodedBinary:           "",
		Args:                          `[{"name": "method", "value": {"string_value": "mint"}},{"name": "address", "value": {"string_value": "friday1qt8k20h3hmdx0qulgpppnlsg92hjjtvn59qkyd"}},{"name": "amount", "value": {"big_int": {"value": "100000", "bit_width": 512}}}]`,
		Fee:                           "10000000",
	}

	body := clictx.Codec.MustMarshalJSON(contractReq)
	req := mustNewRequest(t, "POST", "/contract", bytes.NewReader(body))

	outputBasereq, msgs, err := contractRunMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTContractQuery(t *testing.T) {
	_, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("/contract?data_type=%s&data=%s&path=", "address", "system"), nil)
	res, path, err := getContractQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, path)
}

func TestRESTTransfer(t *testing.T) {
	fromAddr, receipAddr, writer, clictx, basereq := prepare()

	// Body
	transReq := transferReq{
		BaseReq:                    basereq,
		TokenContractAddress:       fromAddr,
		RecipientAddressOrNickname: receipAddr,
		Amount:                     "20000000",
		Fee:                        "10000000",
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
	_, _, writer, clictx, basereq := prepare()

	// Body
	bondReq := bondReq{
		BaseReq: basereq,
		Amount:  "100000000",
		Fee:     "10000000",
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
	_, _, writer, clictx, basereq := prepare()

	// Body
	bondReq := bondReq{
		BaseReq: basereq,
		Amount:  "100000000",
		Fee:     "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(bondReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/unbond", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := bondUnbondMsgCreator(false, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTDelegate(t *testing.T) {
	fromAddr, _, writer, clictx, basereq := prepare()

	// Body
	delegateReq := delegateReq{
		BaseReq:          basereq,
		ValidatorAddress: fromAddr,
		Amount:           "100000000",
		Fee:              "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(delegateReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/delegate", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := delegateUndelegateMsgCreator(true, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTUndelegate(t *testing.T) {
	fromAddr, _, writer, clictx, basereq := prepare()

	// Body
	delegateReq := delegateReq{
		BaseReq:          basereq,
		ValidatorAddress: fromAddr,
		Amount:           "100000000",
		Fee:              "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(delegateReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/undelegate", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := delegateUndelegateMsgCreator(false, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTRedelegate(t *testing.T) {
	srcAddr, destAddr, writer, clictx, basereq := prepare()

	// Body
	delegateReq := redelegateReq{
		BaseReq:              basereq,
		SrcValidatorAddress:  srcAddr,
		DestValidatorAddress: destAddr,
		Amount:               "100000000",
		Fee:                  "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(delegateReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/undelegate", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := delegateUndelegateMsgCreator(false, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTVote(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()
	targetContractAddress := sdk.ContractHashAddress(types.SYSTEM_ACCOUNT)

	// Body
	delegateReq := voteReq{
		BaseReq:                basereq,
		TargetContrractAddress: targetContractAddress.String(),
		Amount:                 "100000000",
		Fee:                    "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(delegateReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/vote", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := voteUnvoteMsgCreator(true, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTUnvote(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()
	targetContractAddress := sdk.ContractHashAddress(types.SYSTEM_ACCOUNT)

	// Body
	delegateReq := voteReq{
		BaseReq:                basereq,
		TargetContrractAddress: targetContractAddress.String(),
		Amount:                 "100000000",
		Fee:                    "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(delegateReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/unvote", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := voteUnvoteMsgCreator(false, writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTClaimReward(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()

	// Body
	claimReq := claimReq{
		BaseReq:            basereq,
		RewardOrCommission: types.RewardValue,
		Amount:             "100000000",
		Fee:                "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(claimReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/claim", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := claimMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputBasereq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTClaimCommission(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()

	// Body
	claimReq := claimReq{
		BaseReq:            basereq,
		RewardOrCommission: types.CommissionValue,
		Amount:             "100000000",
		Fee:                "10000000",
	}

	// http.request
	body := clictx.Codec.MustMarshalJSON(claimReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/claim", types.ModuleName), bytes.NewReader(body))

	outputBasereq, msgs, err := claimMsgCreator(writer, clictx, req)

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

func TestRESTGetDelegatorFromValidator(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/delegator?validator=%s", types.ModuleName, fromAddr), nil)
	res, err := getDelegatorQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetDelegatorFromDelegator(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/delegator?delegator=%s", types.ModuleName, fromAddr), nil)
	res, err := getDelegatorQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetDelegatorFromAll(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/delegator?validator=%s&delegator=%s", types.ModuleName, fromAddr, fromAddr), nil)
	res, err := getDelegatorQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetVoterFromAddress(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/voter?address=%s", types.ModuleName, fromAddr), nil)
	res, err := getVoterQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetVoterFromDapp(t *testing.T) {
	_, _, writer, clictx, _ := prepare()
	hash := hex.EncodeToString(types.SYSTEM_ACCOUNT)

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/voter?hash=%s", types.ModuleName, hash), nil)
	res, err := getVoterQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetVoterFromAll(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()
	hash := hex.EncodeToString(types.SYSTEM_ACCOUNT)

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/voter?hash=%s&address=%s", types.ModuleName, hash, fromAddr), nil)
	res, err := getVoterQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetReward(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/reward?address=%s", types.ModuleName, fromAddr), nil)
	res, err := getVoterQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTGetCommission(t *testing.T) {
	fromAddr, _, writer, clictx, _ := prepare()

	req := mustNewRequest(t, "GET", fmt.Sprintf("%s/commission?address=%s", types.ModuleName, fromAddr), nil)
	res, err := getVoterQuerying(writer, clictx, req)

	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestRESTCreateValidator(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()

	createValidatorReq := createValidatorReq{
		BaseReq:     basereq,
		ConsPubKey:  "fridayvalconspub16jrl8jvqq9k957nfd43n2dnyxc6nsazpgf5yuwtzfe6kku63ga6nvtmcdeg92vj4gy4kkd62vd69vvnhx935w5zpw9ex7733tft8we6evemzke66xv4ks56gfdvx66ndfye5x5z9fs6j74z6g3u4zdzd0p8hw6mr24k8wjzx0ghhz5z8vdm92vjs2e8xwdn5xpvxu56fvejnj7t6wsens5gwxlen9",
		Description: types.NewDescription("moniker", "identity", "https://test.io", "details"),
	}

	body := clictx.Codec.MustMarshalJSON(createValidatorReq)
	req := mustNewRequest(t, "POST", fmt.Sprintf("/%s/validator", types.ModuleName), bytes.NewReader((body)))

	outputCreateValidatorReq, msgs, err := createValidatorMsgCreator(writer, clictx, req)

	require.NoError(t, err)
	require.Equal(t, outputCreateValidatorReq, basereq)
	require.NotNil(t, msgs)
}

func TestRESTEditValidator(t *testing.T) {
	_, _, writer, clictx, basereq := prepare()

	editValidatorReq := editValidatorReq{
		BaseReq:     basereq,
		Description: types.NewDescription("moniker", "identity", "https://test.io", "details"),
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
