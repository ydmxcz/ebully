package ebully

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/ydmxcz/ebully/nodeid"
	"github.com/ydmxcz/ebully/pb/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CharacterType int

type NodeMeta struct {
	info     *node.NodeInfo
	memory   int
	tcp4addr string
	rwLock   sync.RWMutex
	// client   node.NodeClient
}

func NewNodeMeta(nodeInfo *node.NodeInfo) *NodeMeta {
	tcp4addr, memory := nodeid.Decode(nodeInfo.Id)
	nodeMeta := &NodeMeta{
		info:     nodeInfo,
		tcp4addr: strings.TrimRight(string(tcp4addr[:]), "\x00"),
		memory:   memory,
	}
	return nodeMeta
}

func (meta *NodeMeta) UpdateId(id uint64) *NodeMeta {
	meta.rwLock.Lock()
	defer meta.rwLock.Unlock()
	meta.info.Id = id
	tcp4, mem := nodeid.Decode(meta.info.Id)
	meta.tcp4addr = strings.TrimRight(string(tcp4[:]), "\x00")
	meta.memory = mem
	return meta
}

func (meta *NodeMeta) UpdateInfo(nodeInfo *node.NodeInfo) *NodeMeta {
	meta.rwLock.Lock()
	defer meta.rwLock.Unlock()
	meta.info = nodeInfo
	tcp4, mem := nodeid.Decode(meta.info.Id)
	meta.tcp4addr = strings.TrimRight(string(tcp4[:]), "\x00")
	meta.memory = mem
	return meta
}

// Address return the address string
func (meta *NodeMeta) Address() string {
	meta.rwLock.RLock()
	defer meta.rwLock.RUnlock()
	return meta.tcp4addr
}

// Address return the address string
func (meta *NodeMeta) ServiceAddress() string {
	meta.rwLock.RLock()
	defer meta.rwLock.RUnlock()
	return meta.info.ServiceAddress
}

// SendHearbeat to the address of node in the node meta
func (meta *NodeMeta) Leave(nodeInfo *node.LeaveReq) (*node.LeaveResp, error) {
	return grpcClient(meta.Address(), func(nc node.NodeClient) (*node.LeaveResp, error) {
		return nc.LeaveRpc(context.Background(), nodeInfo)
	})
}

// SendHearbeat to the address of node in the node meta
func (meta *NodeMeta) SendHearbeat(nodeInfo *node.NodeInfo) (*node.HearbeatResp, error) {

	return grpcClient(meta.Address(), func(nc node.NodeClient) (*node.HearbeatResp, error) {
		return nc.HearbeatRpc(context.Background(), nodeInfo)
	})
}

func createGrpcClient(addr string) (client node.NodeClient, err error) {
	conn, err := grpc.DialContext(context.Background(), addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return
	}
	return node.NewNodeClient(conn), nil
}

func grpcClient[Resp any](addr string, do func(node.NodeClient) (Resp, error)) (resp Resp, err error) {
	ctx, cf := context.WithTimeout(context.Background(), rpcTimeOut)
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	defer cf()
	if err != nil {
		return
	}
	defer conn.Close()
	return do(node.NewNodeClient(conn))
}

// Meet to the address of node in the node meta
func (meta *NodeMeta) Meet(msg *node.NodeInfo) (*node.MeetResp, error) {
	return grpcClient(meta.Address(), func(nc node.NodeClient) (*node.MeetResp, error) { return nc.MeetRpc(context.Background(), msg) })
}

func (meta *NodeMeta) NodeInfoList() (*node.NodeInfoResp, error) {
	return grpcClient(meta.Address(), func(nc node.NodeClient) (*node.NodeInfoResp, error) {
		return nc.PeerInfosRpc(context.Background(), &node.EmptyMessage{})
	})
}

// UinqueId return the id of node ,it implement the interface named loadbalance.Instance.
func (meta *NodeMeta) InstanceID() uint64 {
	meta.rwLock.RLock()
	defer meta.rwLock.RUnlock()
	return meta.info.Id
}

// Weight return the its memory capacity as the weight ,
// it implement the interface named loadbalance.Instance.
func (meta *NodeMeta) InstanceWeight() int {
	meta.rwLock.RLock()
	defer meta.rwLock.RUnlock()
	return meta.memory
}

type NodeMetaList []*NodeMeta

func (metaList NodeMetaList) Len() int {
	return len(metaList)
}

func (metaList NodeMetaList) Less(i, j int) bool {
	return metaList[i].info.Id < metaList[j].info.Id
}

func (metaList NodeMetaList) Swap(i, j int) {
	metaList[i], metaList[j] = metaList[j], metaList[i]
}

type SafeNodeMetaList struct {
	mutex sync.Mutex
	NodeMetaList
}

func (so *SafeNodeMetaList) Sort() {
	sort.Sort(so.NodeMetaList)
}

func (so *SafeNodeMetaList) BinarySearch(target uint64) int {
	return sort.Search(len(so.NodeMetaList),
		func(i int) bool {
			return so.NodeMetaList[i].info.Id >= target
		})
}

func (so *SafeNodeMetaList) DeleteIndex(i int) {
	so.NodeMetaList = append(so.NodeMetaList[:i], so.NodeMetaList[i+1:]...)
}

func (so *SafeNodeMetaList) Append(node *NodeMeta) {
	so.NodeMetaList = append(so.NodeMetaList, node)
}

func asyncSendNodeMessage(req *node.MessageReq, meta *NodeMeta) {
	// send async
	go func() {
		ctx, cf := context.WithTimeout(context.Background(), rpcTimeOut)
		conn, err := grpc.DialContext(ctx, meta.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())
		defer cf()
		if err != nil {
			log.Fatalln(err)
		}
		defer conn.Close()
		node.NewNodeClient(conn).NodeMessageRpc(context.Background(), req)
	}()
}

func (so *SafeNodeMetaList) Broadcast(msg *node.MessageReq) {
	for i := 0; i < len(so.NodeMetaList); i++ {
		asyncSendNodeMessage(msg, so.NodeMetaList[i])
	}
}

func (so *SafeNodeMetaList) SafeDo(f func(so *SafeNodeMetaList)) {
	so.mutex.Lock()
	defer so.mutex.Unlock()
	f(so)
}
