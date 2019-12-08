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
