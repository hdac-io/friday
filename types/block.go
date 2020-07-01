package types

import (
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
)

type CandidateBlock struct {
	Hash            []byte                       `json:"hash"`
	State           []byte                       `json:"state"`
	Bonds           []*ipc.Bond                  `json:"bonds"`
	Deploys         []*ipc.DeployItem            `json:"deploys"`
	Effects         []*transforms.TransformEntry `json:"effects"`
	ProtocolVersion *state.ProtocolVersion       `json:"protocol_version"`
}
