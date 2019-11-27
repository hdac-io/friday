package types

import (
	sdk "github.com/hdac-io/friday/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodePublicKeyDecode      sdk.CodeType = 101
	CodeProtocolVersionParse sdk.CodeType = 102
	CodeInvalidWasmPath      sdk.CodeType = 103
)

// ErrPublicKeyDecode :
func ErrPublicKeyDecode(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(
		codespace, CodePublicKeyDecode, "Could not decode public key as Base64")
}

// ErrProtocolVersionParse is an error
func ErrProtocolVersionParse(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(
		codespace, CodeProtocolVersionParse,
		"Error occurs in parsing protocol version in chainspec")
}

// ErrInvalidWasmPath is an error
func ErrInvalidWasmPath(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(
		codespace, CodeInvalidWasmPath, "Invalid wasm path in chainspec")
}
