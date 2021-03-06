package executionlayer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/grpc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/storedvalue"
	"github.com/hdac-io/friday/codec"
	sdk "github.com/hdac-io/friday/types"
	abci "github.com/hdac-io/tendermint/abci/types"

	"github.com/hdac-io/friday/x/executionlayer/types"
)

const (
	QueryEEDetail        = "querydetail"
	QueryEEBalanceDetail = "querybalancedetail"
	QueryStakeDetail     = "querystakedetail"
	QueryVoteDetail      = "queryvotedetail"

	QueryValidator    = "queryvalidator"
	QueryAllValidator = "queryallvalidator"

	QueryDelegator = "querydelegator"
	QueryVoter     = "queryvoter"

	QueryReward     = "queryreward"
	QueryCommission = "querycommission"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper ExecutionLayerKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryEEDetail:
			return queryEEDetail(ctx, path[1:], req, keeper)
		case QueryEEBalanceDetail:
			return queryBalanceDetail(ctx, path[1:], req, keeper)
		case QueryStakeDetail:
			return queryStakeDetail(ctx, req, keeper)
		case QueryVoteDetail:
			return queryVoteDetail(ctx, req, keeper)
		case QueryValidator:
			return queryValidator(ctx, req, keeper)
		case QueryAllValidator:
			return queryAllValidator(ctx, req, keeper)
		case QueryDelegator:
			return queryDelegator(ctx, req, keeper)
		case QueryVoter:
			return queryVoter(ctx, req, keeper)
		case QueryReward:
			return queryReward(ctx, req, keeper)
		case QueryCommission:
			return queryCommission(ctx, req, keeper)
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

	ctx = ctx.WithBlockHeight(req.Height)
	res, err := getQueryResult(ctx, keeper, param.KeyType, param.KeyData, param.Path)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, err.Error())
	}

	return res, nil
}

func queryBalanceDetail(ctx sdk.Context, path []string, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryGetBalanceDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeState := keeper.GetUnitHashMap(ctx, req.GetHeight()).EEState
	protocolVersion := keeper.GetProtocolVersion(ctx)
	val, errMsg := grpc.QueryBalance(keeper.client, eeState, param.Address, &protocolVersion)
	if errMsg != "" {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
	}

	queryvalue := &state.Value{Value: &state.Value_StringValue{StringValue: val}}

	jsonMarshaler := jsonpb.Marshaler{}
	res := &bytes.Buffer{}
	err = jsonMarshaler.Marshal(res, queryvalue)

	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}
	return res.Bytes(), nil
}

func queryStakeDetail(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryGetStakeDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeState := keeper.GetUnitHashMap(ctx, req.GetHeight()).EEState
	protocolVersion := keeper.GetProtocolVersion(ctx)
	val, errMsg := grpc.QueryStake(keeper.client, eeState, param.Address, &protocolVersion)
	if errMsg != "" {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
	}

	queryvalue := &state.Value{Value: &state.Value_StringValue{StringValue: val}}

	jsonMarshaler := jsonpb.Marshaler{}
	res := &bytes.Buffer{}
	err = jsonMarshaler.Marshal(res, queryvalue)

	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}
	return res.Bytes(), nil
}

func queryVoteDetail(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryGetVoteDetail
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeState := keeper.GetUnitHashMap(ctx, req.GetHeight()).EEState
	protocolVersion := keeper.GetProtocolVersion(ctx)

	val := ""
	errMsg := ""
	if !param.Address.Empty() {
		val, errMsg = grpc.QueryVoting(keeper.client, eeState, param.Address, &protocolVersion)
		if errMsg != "" {
			return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
		}
	} else if param.Dapp != "" {
		var key storedvalue.Key
		if strings.HasPrefix(param.Dapp, sdk.Bech32PrefixContractURef) {
			urefAddress, err := sdk.ContractUrefAddressFromBech32(param.Dapp)
			if err != nil {
				return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
			}

			uref := storedvalue.NewURef(urefAddress.Bytes(), state.Key_URef_NONE)
			key = storedvalue.NewKeyFromURef(uref)
		} else if strings.HasPrefix(param.Dapp, sdk.Bech32PrefixContractHash) {
			hash, err := sdk.ContractHashAddressFromBech32(param.Dapp)
			if err != nil {
				return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
			}

			key = storedvalue.NewKeyFromHash(hash.Bytes())

		}
		val, errMsg = grpc.QueryVoted(keeper.client, eeState, key.ToBytes(), &protocolVersion)
		if errMsg != "" {
			return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
		}
	}

	queryvalue := &state.Value{Value: &state.Value_StringValue{StringValue: val}}

	jsonMarshaler := jsonpb.Marshaler{}
	res := &bytes.Buffer{}
	err = jsonMarshaler.Marshal(res, queryvalue)

	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}
	return res.Bytes(), nil
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

	ctx = ctx.WithBlockHeight(req.Height)
	res, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, types.PosContractName)
	var storedValue storedvalue.StoredValue
	storedValue, err, _ = storedValue.FromBytes(res)
	if err != nil {
		return []byte{}, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	validator.Stake = storedValue.Contract.NamedKeys.GetValidatorStake(param.ValidatorAddr)

	res, err = codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryAllValidator(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	validators := keeper.GetAllValidators(ctx)

	ctx = ctx.WithBlockHeight(req.Height)
	res, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, types.PosContractName)
	var storedValue storedvalue.StoredValue
	storedValue, err, _ = storedValue.FromBytes(res)
	if err != nil {
		return []byte{}, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeValidators := storedValue.Contract.NamedKeys.GetAllValidators()

	for _, validator := range validators {
		valEEAddrStr := hex.EncodeToString(validator.OperatorAddress)
		validator.Stake = eeValidators[valEEAddrStr]
	}

	res, err = codec.MarshalJSONIndent(types.ModuleCdc, validators)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

// GetQueryResult queries with whole parameters
func getQueryResult(ctx sdk.Context, k ExecutionLayerKeeper,
	keyType string, keyData string, path string) ([]byte, error) {
	arrPath := []string{}
	if path != "" {
		arrPath = strings.Split(path, "/")
	}

	protocolVersion := k.GetProtocolVersion(ctx)
	stateHash := k.GetUnitHashMap(ctx, ctx.BlockHeight()).EEState
	if len(stateHash) == 0 {
		stateHash = ctx.CandidateBlock().State
	}
	keyDataBytes, err := toBytes(keyType, keyData, k.NicknameKeeper, ctx)
	if err != nil {
		return []byte{}, err
	}
	res, errstr := grpc.Query(k.client, stateHash, keyType, keyDataBytes, arrPath, &protocolVersion)
	if errstr != "" {
		return []byte{}, fmt.Errorf(errstr)
	}

	return res, nil
}

func queryDelegator(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param QueryDelegatorParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	ctx = ctx.WithBlockHeight(req.Height)
	res, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, types.PosContractName)
	var storedValue storedvalue.StoredValue
	storedValue, err, _ = storedValue.FromBytes(res)
	if err != nil {
		return []byte{}, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	var resMap map[string]string
	if !param.ValidatorAddr.Empty() {
		resMap = storedValue.Contract.NamedKeys.GetDelegateFromValidator(param.ValidatorAddr)

		if !param.DelegatorAddr.Empty() {
			delegateAddressStr := hex.EncodeToString(param.DelegatorAddr)
			resMap = map[string]string{delegateAddressStr: resMap[delegateAddressStr]}
		}
	}
	if !param.DelegatorAddr.Empty() {
		resMap = storedValue.Contract.NamedKeys.GetDelegateFromDelegator(param.DelegatorAddr)

		if !param.ValidatorAddr.Empty() {
			validatorAddressStr := hex.EncodeToString(param.ValidatorAddr)
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
			Address: address,
			Amount:  amount,
		}
		delegators = append(delegators, delegator)
	}

	res, err = codec.MarshalJSONIndent(types.ModuleCdc, delegators)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryVoter(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var paramUref QueryVoterParamsUref
	var paramHash QueryVoterParamsHash
	var param QueryVoterParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &paramUref)
	var contractKey storedvalue.Key
	if err != nil {
		err = types.ModuleCdc.UnmarshalJSON(req.Data, &paramHash)
		if err != nil {
			return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
		}
		param = paramHash
		contractKey = storedvalue.NewKeyFromHash(param.GetContract().Bytes())
	} else {
		param = paramUref
		uref := storedvalue.NewURef(param.GetContract().Bytes(), state.Key_URef_NONE)
		contractKey = storedvalue.NewKeyFromURef(uref)
	}

	ctx = ctx.WithBlockHeight(req.Height)
	res, err := getQueryResult(ctx, keeper, types.ADDRESS, types.SYSTEM, types.PosContractName)
	var storedValue storedvalue.StoredValue
	storedValue, err, _ = storedValue.FromBytes(res)
	if err != nil {
		return []byte{}, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}
	var resMap map[string]string
	if !param.GetAddress().Empty() {
		resMap = storedValue.Contract.NamedKeys.GetVotingDappFromUser(param.GetAddress())

		if !param.GetContract().Empty() {
			dappAddressHex := hex.EncodeToString(param.GetContract().Bytes())
			resMap = map[string]string{dappAddressHex: resMap[dappAddressHex]}
		}
	}
	if !param.GetContract().Empty() {
		resMap = storedValue.Contract.NamedKeys.GetVotingUserFromDapp(contractKey.ToBytes())

		if !param.GetAddress().Empty() {
			addressHex := hex.EncodeToString(param.GetAddress())
			resMap = map[string]string{addressHex: resMap[addressHex]}
		}
	}

	voters := []types.QueryVoterResponse{}
	for addressStr, amount := range resMap {
		addressBytes, err := hex.DecodeString(addressStr)
		if err != nil {
			return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeInvalidAddress, "Can't convert address {}")
		}
		var address sdk.Address
		if len(addressBytes) == sdk.AddrLen {
			address = sdk.AccAddress(addressBytes)
		} else {
			var key storedvalue.Key
			key, err, _ := key.FromBytes(addressBytes)
			if err != nil {
				return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeInvalidAddress, "Can't convert address {}")
			}

			switch key.KeyID {
			case storedvalue.KEY_ID_HASH:
				address = sdk.ContractHashAddress(key.Hash)
			case storedvalue.KEY_ID_UREF:
				address = sdk.ContractUrefAddress(key.Uref.Address)
			case storedvalue.KEY_ID_ACCOUNT:
				address = sdk.AccAddress(key.Account.PublicKey)
			case storedvalue.KEY_ID_LOCAL:
			default:
				return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeInvalidAddress, "Can't convert key")
			}
		}

		voter := types.NewQueryVoterResponse(address.String(), amount)
		voters = append(voters, voter)
	}

	res, err = codec.MarshalJSONIndent(types.ModuleCdc, voters)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return res, nil
}

func queryReward(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param types.QueryGetReward
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeState := keeper.GetUnitHashMap(ctx, req.GetHeight()).EEState
	protocolVersion := keeper.GetProtocolVersion(ctx)
	val, errMsg := grpc.QueryReward(keeper.client, eeState, param.Address, &protocolVersion)
	if errMsg != "" {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
	}

	queryvalue := &state.Value{Value: &state.Value_StringValue{StringValue: val}}

	jsonMarshaler := jsonpb.Marshaler{}
	res := &bytes.Buffer{}
	err = jsonMarshaler.Marshal(res, queryvalue)

	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}
	return res.Bytes(), nil
}

func queryCommission(ctx sdk.Context, req abci.RequestQuery, keeper ExecutionLayerKeeper) ([]byte, sdk.Error) {
	var param types.QueryGetCommission
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &param)
	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}

	eeState := keeper.GetUnitHashMap(ctx, req.GetHeight()).EEState
	protocolVersion := keeper.GetProtocolVersion(ctx)
	val, errMsg := grpc.QueryCommission(keeper.client, eeState, param.Address, &protocolVersion)
	if errMsg != "" {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", errMsg)
	}

	queryvalue := &state.Value{Value: &state.Value_StringValue{StringValue: val}}

	jsonMarshaler := jsonpb.Marshaler{}
	res := &bytes.Buffer{}
	err = jsonMarshaler.Marshal(res, queryvalue)

	if err != nil {
		return nil, sdk.NewError(sdk.CodespaceUndefined, sdk.CodeUnknownRequest, "Bad request: {}", err.Error())
	}
	return res.Bytes(), nil
}
