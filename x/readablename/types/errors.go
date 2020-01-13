package types

import (
	sdk "github.com/hdac-io/friday/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeBadQueryRequest        sdk.CodeType = 400
	CodeNoRegisteredReadableID sdk.CodeType = 404
)

// ErrBadQueryRequest - malform query request
func ErrBadQueryRequest(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeBadQueryRequest, "bad query request")
}

// ErrNoRegisteredReadableID - no registered readable name
func ErrNoRegisteredReadableID(codespace sdk.CodespaceType, readableid string) sdk.Error {
	return sdk.NewError(codespace, CodeNoRegisteredReadableID, "no registered readable name: %v", readableid)
}
