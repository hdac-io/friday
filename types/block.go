package types

import (
	"sync"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/casper/consensus/state"
	"github.com/hdac-io/casperlabs-ee-grpc-go-util/protobuf/io/casperlabs/ipc"
)

type CandidateBlock struct {
	Hash            []byte                 `json:"hash"`
	State           []byte                 `json:"state"`
	Bonds           []*ipc.Bond            `json:"bonds"`
	ProtocolVersion *state.ProtocolVersion `json:"protocol_version"`
	TxsCount        int64                  `json:"txs_count"`
	WaitGroup       sync.WaitGroup         `json:"wait_group"`
	DeployPQueue    *queue.PriorityQueue   `json:"deploy_priority_queue"`
	AnteCond        *sync.Cond             `json:"anti_condition"`
	CurrentTxIndex  int                    `json:"current_tx_index"`
}

type ItemDeploy struct {
	TxIndex    int             `json:"tx_index"`
	MsgIndex   int             `json:"msg_index"`
	Deploy     *ipc.DeployItem `json:"deploy"`
	LogChannel chan string     `json:"deploy_channel"`
}

func (i ItemDeploy) Compare(src queue.Item) int {
	srcDep := src.(*ItemDeploy)

	if i.TxIndex == srcDep.TxIndex {
		if i.MsgIndex > srcDep.MsgIndex {
			return 1
		} else if i.MsgIndex < srcDep.MsgIndex {
			return -1
		} else {
			return 0
		}
	} else if i.TxIndex > srcDep.TxIndex {
		return 1
	} else if i.TxIndex < srcDep.TxIndex {
		return -1
	}

	return -1
}
