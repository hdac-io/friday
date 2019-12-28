package types

import (
	"fmt"
	sdk "github.com/hdac-io/friday/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodePublicKeyDecode      sdk.CodeType = 101
	CodeProtocolVersionParse sdk.CodeType = 102
	CodeTomlParse            sdk.CodeType = 103
	CodeInvalidValidator     sdk.CodeType = 201
	CodeInvalidDelegation    sdk.CodeType = 202
	CodeInvalidInput         sdk.CodeType = 203
	CodeInvalidAddress		 sdk.CodeType = sdk.CodeInvalidAddress
)

// ErrPublicKeyDecode is an error
func ErrPublicKeyDecode(codespace sdk.CodespaceType, publicKey string) sdk.Error {
	return sdk.NewError(
		codespace, CodePublicKeyDecode, "Could not decode public key as Base64 : %v", publicKey)
}

// ErrProtocolVersionParse is an error
func ErrProtocolVersionParse(codespace sdk.CodespaceType, protocolVersion string) sdk.Error {
	return sdk.NewError(
		codespace, CodeProtocolVersionParse,
		"Could not parse Protocol Version : %v", protocolVersion)
}

// ErrTomlParse is an error
func ErrTomlParse(codespace sdk.CodespaceType, keyString string) sdk.Error {
	return sdk.NewError(
		codespace, CodeTomlParse,
		"Could not parse Toml with : %v", keyString)
}

func ErrNilValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "validator address is nil")
}

func ErrBadValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAddress, "validator address is invalid")
}

func ErrDescriptionLength(codespace sdk.CodespaceType, descriptor string, got, max int) sdk.Error {
	msg := fmt.Sprintf("bad description length for %v, got length %v, max is %v", descriptor, got, max)
	return sdk.NewError(codespace, CodeInvalidValidator, msg)
}

func ErrNilDelegatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "delegator address is nil")
}

func ErrBadDelegationAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "unexpected address length for this (address, validator) pair")
}

func ErrBadDelegationAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "amount must be > 0")
}
