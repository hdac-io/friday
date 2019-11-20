package types

import (
	"fmt"
)

// QueryExecutionLayer payload for a EE query
type QueryExecutionLayerDetail struct {
	StateHash []byte `json:"state_hash"`
	KeyType   string `json:"key_type"`
	KeyData   []byte `json:"key_data"`
	Path      string `json:"path"`
}

// implement fmt.Stringer
func (q QueryExecutionLayerDetail) String() string {
	return fmt.Sprintf("State: %s\nKey type: %s\nKey data: %s\nPath: %s", q.StateHash, q.KeyType, q.KeyData, q.Path)
}

// QueryExecutionLayer payload for a EE query
type QueryExecutionLayer struct {
	KeyType string `json:"key_type"`
	KeyData []byte `json:"key_data"`
	Path    string `json:"path"`
}

// implement fmt.Stringer
func (q QueryExecutionLayer) String() string {
	return fmt.Sprintf("Key type: %s\nKey data: %s\nPath: %s", q.KeyType, q.KeyData, q.Path)
}

// QueryGetBalanceDetail payload for balance query
type QueryGetBalanceDetail struct {
	StateHash []byte
	Address   string
}

// implement fmt.Stringer
func (q QueryGetBalanceDetail) String() string {
	return fmt.Sprintf("State: %s\nAddress: %s", q.StateHash, q.Address)
}

// QueryGetBalance payload for balance query in the latest data
type QueryGetBalance struct {
	Address string
}

// implement fmt.Stringer
func (q QueryGetBalance) String() string {
	return fmt.Sprintf("Address: %s", q.Address)
}

// QueryExecutionLayerResp is used for response of EE query
type QueryExecutionLayerResp struct {
	Value string `json:"value"`
}
