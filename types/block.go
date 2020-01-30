package types

import (
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc/transforms"
)

type CandidateBlock struct {
	Hash    []byte                       `json:"hash"`
	State   []byte                       `json:"state"`
	Effects []*transforms.TransformEntry `json:"effects`
}
