package ebully

import (
	"errors"
	"sort"

	"github.com/ydmxcz/ebully/nodeid"
	nodepb "github.com/ydmxcz/ebully/pb/node"
)

type NodeCharacterType int

func (nct NodeCharacterType) String() string {
	if nct == Master {
		return "master"
	} else {
		return "follower"
	}
}

type EBully struct {
	node *Node
}

func NewEBully(cfg Config) *EBully {
	return &EBully{
		node: newNode(cfg),
	}
}

func (e *EBully) Weight() int {
	return nodeid.Weight(e.ID())
}

func (e *EBully) ID() uint64 { return e.node.selfInfo.Id }

func (e *EBully) Start() { e.node.Start() }

func (e *EBully) PeerList() []*nodepb.NodeInfo {
	pi := e.node.PeerInfos()
	sort.Slice(pi, func(i, j int) bool {
		return pi[i].Id > pi[j].Id
	})
	return pi
}

func (e *EBully) PeerIds() []uint64 {
	return e.node.PeerIds()
}

var (
	errMasterIsItself = errors.New("target master node is self")
)

func (e *EBully) Meet(masterAddr string) (err error) {
	tcp4, _ := nodeid.Decode(e.node.selfInfo.Id)
	if masterAddr == string(tcp4[:]) {
		return errMasterIsItself
	}
	defer func() {
		if v := recover(); v != nil {
			err = v.(error)
		}
	}()
	e.node.masterMeta = &NodeMeta{
		info: &nodepb.NodeInfo{
			Id:        nodeid.Encode(0, masterAddr),
			Character: Master,
		},
		tcp4addr: masterAddr,
		memory:   0,
	}
	e.node.becomeFollower()
	return err
}

func (e *EBully) Leave() error {
	return e.node.Leave()
}

func (e *EBully) Shutdown() error {
	return e.node.Stop()
}

func (e *EBully) Character() NodeCharacterType {
	return NodeCharacterType(e.node.GetCharacter())
}
