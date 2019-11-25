package types

import (
	sdk "github.com/hdac-io/friday/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeMalforemdAccountsCsv sdk.CodeType = 101
	CodeProtocolVersionParse sdk.CodeType = 102
	CodeInvalidWasmPath      sdk.CodeType = 103
)

// ErrMalforemdAccountsCsv is an error
func ErrMalforemdAccountsCsv(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(
		codespace, CodeMalforemdAccountsCsv, "Malformed account.csv")
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
