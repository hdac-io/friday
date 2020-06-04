package types

import (
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
)

type CandidateBlock struct {
	Hash            []byte      `json:"hash"`
	State           []byte      `json:"state"`
	Bonds           []*ipc.Bond `json:"bonds"`
	ProtocolVersion *state.ProtocolVersion
}
