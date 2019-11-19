package types

import (
	"fmt"
)

// QueryExecutionLayer payload for a UnitAccount query
type QueryExecutionLayer struct {
	StateHash []byte `json:"state_hash"`
	KeyType   string `json:"key_type"`
	KeyData   []byte `json:"key_data"`
	Path      string `json:"path"`
}

// implement fmt.Stringer
func (q QueryExecutionLayer) String() string {
	return fmt.Sprintf("State: %s\nKey type: %s\nKey data: %s\nPath: %s", q.StateHash, q.KeyType, q.KeyData, q.Path)
}
