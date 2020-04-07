package executionlayer

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"

	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	QueryEEDetail        = "querydetail"
	QueryEEBalanceDetail = "querybalancedetail"

	QueryValidator    = "queryvalidator"
	QueryAllValidator = "queryallvalidator"

	QueryDelegator = "querydelegator"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper ExecutionLayerKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryEEDetail:
			return queryEEDetail(ctx, path[1:], req, keeper)
		case QueryEEBalanceDetail:
			return queryBalanceDetail(ctx, path[1:], req, keeper)
		case QueryValidator:
			return queryValidator(ctx, req, keeper)
		case QueryAllValidator:
			return queryAllValidator(ctx, keeper)
		case QueryDelegator:
			return queryDelegator(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown ee query")
		}
	}
}

func queryEEDetail(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryExecutionLayerDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	storedValue, err := getQueryResult(ctx, keeper, param.BlockHash, param.KeyType, param.KeyData, param.Path)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, err.Error())
	}

	var valueStr string
	switch storedValue.Type {
	case storedvalue.TYPE_ACCOUNT:
		valueStr = storedValue.Account.ToStateValue().String()
	case storedvalue.TYPE_CONTRACT:
		valueStr = storedValue.Contract.ToStateValue().String()
	case storedvalue.TYPE_CL_VALUE:
		valueStr = storedValue.ClValue.ToStateValues().String()
	}

	qryvalue := QueryExecutionLayerResp{
		Value: valueStr,
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, qryvalue)
	return res, nil
}

func queryBalanceDetail(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryGetBalanceDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	var blockHash []byte
	if param.BlockHash != "" {
		blockHash, err = hex.DecodeString(param.BlockHash)
		if err != nil {
			return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
		}
	} else {
		blockHash = ctx.BlockHeader().LastBlockId.Hash
	}

	eeState := keeper.GetEEState(ctx, blockHash)
	protocolVersion := keeper.MustGetProtocolVersion(ctx)
	val, errMsg := grpc.QueryBalance(keeper.client, eeState, param.Address.ToEEAddress(), &protocolVersion)
	if errMsg != "" {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
	}

	queryvalue := QueryExecutionLayerResp{
		Value: val,
	}

	res, _ := codec.MarshalJSONIndent(keeper.cdc, queryvalue)
	return res, nil
}

func queryValidator(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryValidatorParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	validator, found := keeper.GetValidator(ctx, param.ValidatorAddr)
	if !found {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, err.Error())
	}

	storedValue, err := getQueryResult(ctx, keeper, "", types.ADDRESS, types.SYSTEM, types.PosContractName)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	validator.Stake = storedValue.Contract.NamedKeys.GetValidatorStake(param.ValidatorAddr.ToEEAddress())

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryAllValidator(ctx sdk.Context, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	validators := keeper.GetAllValidators(ctx)

	storedValue, err := getQueryResult(ctx, keeper, "", types.ADDRESS, types.SYSTEM, types.PosContractName)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeValidators := storedValue.Contract.NamedKeys.GetAllValidators()

	for _, validator := range validators {
		valEEAddrStr := hex.EncodeToString(validator.OperatorAddress.ToEEAddress())
		validator.Stake = eeValidators[valEEAddrStr]
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validators)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

// GetQueryResult queries with whole parameters
func getQueryResult(ctx sdk.Context, k ExecutionLayerKeeper,
	blockhashStr string, keyType string, keyData string, path string) (storedvalue.StoredValue, error) {
	arrPath := strings.Split(path, "/")

	var blockhash []byte
	var err error
	if blockhashStr == "" {
		blockhash = ctx.BlockHeader().LastBlockId.Hash
	} else {
		blockhash, err = hex.DecodeString(blockhashStr)
		if err != nil {
			return storedvalue.StoredValue{}, err
		}
	}

	protocolVersion := k.MustGetProtocolVersion(ctx)
	unitHash := k.GetUnitHashMap(ctx, blockhash)
	keyDataBytes, err := toBytes(keyType, keyData, k.NicknameKeeper, ctx)
	if err != nil {
		return storedvalue.StoredValue{}, err
	}
	res, errstr := grpc.Query(k.client, unitHash.EEState, keyType, keyDataBytes, arrPath, &protocolVersion)
	if errstr != "" {
		return storedvalue.StoredValue{}, fmt.Errorf(errstr)
	}

	var sValue storedvalue.StoredValue
	sValue, err, _ = sValue.FromBytes(res)
	if err != nil {
		return storedvalue.StoredValue{}, err
	}

	return sValue, nil
}

func queryDelegator(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryDelegatorParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	storedValue, err := getQueryResult(ctx, keeper, "", types.ADDRESS, types.SYSTEM, types.PosContractName)

	var resMap map[string]string
	if !param.ValidatorAddr.Empty() {
		resMap = storedValue.Contract.NamedKeys.GetDelegateFromValidator(param.ValidatorAddr.ToEEAddress())

		if !param.DelegatorAddr.Empty() {
			delegateAddressStr := hex.EncodeToString(param.DelegatorAddr.ToEEAddress())
			resMap = map[string]string{delegateAddressStr: resMap[delegateAddressStr]}
		}
	}
	if !param.DelegatorAddr.Empty() {
		resMap = storedValue.Contract.NamedKeys.GetDelegateFromDelegator(param.DelegatorAddr.ToEEAddress())

		if !param.ValidatorAddr.Empty() {
			validatorAddressStr := hex.EncodeToString(param.ValidatorAddr.ToEEAddress())
			resMap = map[string]string{validatorAddressStr: resMap[validatorAddressStr]}
		}
	}

	delegators := types.Delegators{}
	for addressStr, amount := range resMap {
		address, err := hex.DecodeString(addressStr)
		if err != nil {
			return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeInvalidAddress, "Can't convert address {}")
		}
		delegator := types.Delegator{
			Address: sdk.EEAddress(address).ToAccAddress(),
			Amount:  amount,
		}
		delegators = append(delegators, delegator)
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, delegators)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}
