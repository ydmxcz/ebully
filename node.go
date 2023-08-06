package ebully

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync/atomic"
	"time"

	"github.com/alphadose/haxmap"
	nodepb "github.com/ydmxcz/ebully/node"
	"github.com/ydmxcz/loadbalance"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	TemporaryCharacter = iota
	Master
	Follower
)

const (
	NodeUp = iota
	NodeDown
	MasterChange
)

const (
	messageSplitSymbol = ','
	messageEnd         = '.'
	MaxInt64           = 1<<64 - 1
)

// ./judger --enable-cluster=true --http-addr=192.168.0.192:2345 --grpc-addr=192.168.0.192:2346 --tcp-proxy-addr=192.168.0.192:2344 --node-weight=4 -enable-debug=true
// ./judger --enable-cluster=true --http-addr=192.168.0.192:3345 --grpc-addr=192.168.0.192:3346 --tcp-proxy-addr=192.168.0.192:3344 --master-addr=192.168.0.192:2346 --node-weight=3 -enable-debug=true
// ./judger --enable-cluster=true --http-addr=192.168.0.192:4345 --grpc-addr=192.168.0.192:4346 --tcp-proxy-addr=192.168.0.192:4344 --master-addr=192.168.0.192:2346 --node-weight=2 -enable-debug=true
// ./judger --enable-cluster=true --http-addr=192.168.0.192:5345 --grpc-addr=192.168.0.192:5346 --tcp-proxy-addr=192.168.0.192:5344 --master-addr=192.168.0.192:2346 --node-weight=1 -enable-debug=true
var rpcTimeOut time.Duration = time.Second

type Config struct {
	ID             uint64
	MasterAddr     string
	SelfAddr       string
	RetryCount     int
	RpcTimeOut     time.Duration
	HeartBeatTime  time.Duration
	Logger         *zap.Logger
	Listener       net.Listener
	ServiceAddress string

	EnableTcpProxy   bool
	NodeUpCallback   func(*nodepb.NodeInfo)
	NodeDownCallback func(*nodepb.NodeInfo)
}

type Node struct {
	nodepb.UnsafeNodeServer

	ctx       context.Context
	cancleCtx context.CancelFunc

	selfInfo    *nodepb.NodeInfo
	hearbeatMap *haxmap.Map[uint64, int64]
	selfAddr    string
	masterMeta  *NodeMeta
	// follower node hearbeat time
	heartbeatTime time.Duration
	// node reconnect count
	retryCount int

	loadbalancer *loadbalance.DynamicWeighted[uint64, *NodeMeta]

	grpcServer     *grpc.Server
	logger         *zap.Logger
	listener       net.Listener
	enableTcpProxy bool

	nodeUpCallback   func(*nodepb.NodeInfo)
	nodeDownCallback func(*nodepb.NodeInfo)
}

func newNode(cfg Config) *Node {
	rpcTimeOut = cfg.RpcTimeOut
	n := &Node{
		selfInfo: &nodepb.NodeInfo{
			Id:             cfg.ID,
			StartTime:      getTime(),
			ServiceAddress: cfg.ServiceAddress,
		},
		retryCount:    cfg.RetryCount,
		loadbalancer:  loadbalance.NewWeightedDoubleQueue[uint64, *NodeMeta](),
		logger:        cfg.Logger,
		selfAddr:      cfg.SelfAddr,
		heartbeatTime: cfg.HeartBeatTime,
		hearbeatMap:   haxmap.New[uint64, int64](8),
		listener:      cfg.Listener,
	}
	if cfg.MasterAddr != "" {
		n.masterMeta = NewNodeMeta(&nodepb.NodeInfo{
			Id: EncodeNodeID(0, cfg.MasterAddr),
		})
	}
	if n.heartbeatTime == 0 {
		n.heartbeatTime = (time.Second * 5)
	}
	if cfg.RpcTimeOut == 0 {
		rpcTimeOut = time.Second
	} else {
		rpcTimeOut = cfg.RpcTimeOut
	}
	return n
}

func (n *Node) Stop() error {
	err := n.Leave()
	if err != nil {
		return err
	}
	err = n.listener.Close()
	if err != nil {
		return err
	}
	n.grpcServer.Stop()
	return nil
}

func (n *Node) startTcpProxyServer() error {
	n.logger.Info("tcp reverse proxy Listening:", zap.String("address:", n.listener.Addr().String()))
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			return err
		}
		go n.reverseProxy(conn)
	}
}

func (n *Node) reverseProxy(conn net.Conn) {
	// 处理进来的连接
	defer conn.Close()
	// 负载均衡
	ins := n.loadbalancer.Select()
	if ins == nil {
		return
	}
	n.logger.Sugar().Debugln(n.loadbalancer.Size(), ins)
	fconn, err := net.Dial("tcp", ins.ServiceAddress())
	if err != nil {
		return
	}
	defer fconn.Close()
	go io.Copy(conn, fconn)
	io.Copy(fconn, conn)
	n.logger.Sugar().Debugln("conn close")
}

// 获取角色
func (n *Node) GetCharacter() int64 {
	return atomic.LoadInt64(&n.selfInfo.Character)
}

// 获取角色
func (n *Node) SetCharacter(c int64) {
	atomic.StoreInt64(&n.selfInfo.Character, c)
}

// startGrpcServer 启动grpc server
func (n *Node) startGrpcServer() {
	lis, err := net.Listen("tcp", n.selfAddr)
	if err != nil {
		panic(err)
	}
	n.grpcServer = grpc.NewServer()
	// 注册服务
	nodepb.RegisterNodeServer(n.grpcServer, n)
	// 起服务
	go func() {
		n.logger.Sugar().Info("gRPC server is running,address:", n.selfAddr)
		if err := n.grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func (n *Node) Start() {
	n.start()
}

func (n *Node) start() {
	n.startGrpcServer()
	// no master node meta
	if n.masterMeta == nil {
		n.becomeMaster()
	} else { // follower
		n.becomeFollower()
	}
	if n.enableTcpProxy {
		if len(n.selfInfo.ServiceAddress) == 0 {
			panic("enable tcp reverse proxy ,but no service address")
		}
		if n.listener == nil {
			panic("enable tcp revsese proxy,but listener is nil")
		}
		go n.startTcpProxyServer()
	}
	// n.logger.Debug("", zap.Uint64("node id", n.selfInfo.Id))
}

// changeFollerCharacter 切换成主节点角色
func (n *Node) becomeMaster() {
	if n.GetCharacter() == Master {
		return
	}
	if n.ctx != nil {
		n.cancleCtx()
	}
	n.ctx, n.cancleCtx = context.WithCancel(context.Background())
	n.SetCharacter(Master)
	n.loadbalancer.Add(NewNodeMeta(n.selfInfo))
	n.hearbeatMap.Set(n.selfInfo.Id, 0)
	go n.checkFollowerHearbeat()
	n.logger.Debug("node become master")
}

func (n *Node) SetMasterMeta(nodeinfo *nodepb.NodeInfo) *Node {
	n.masterMeta.UpdateInfo(nodeinfo)
	return n
}

// becomeFollower 切换成从节点角色
func (n *Node) becomeFollower() {
	if n.GetCharacter() == Follower {
		return
	}
	if n.ctx != nil {
		n.cancleCtx()
	}
	n.ctx, n.cancleCtx = context.WithCancel(context.Background())
	n.SetCharacter(Follower)
	for {
		// 尝试去链接启动时指定的地址
		mr, err := n.masterMeta.Meet(n.selfInfo)
		if err != nil {
			n.logger.Error(err.Error())
			n.becomeMaster()
			return
		}
		// 如果返回的重定向地址与一开始指定的地址不一样，则执行重定向
		if mr.Redirect.Id != n.masterMeta.InstanceID() {
			n.masterMeta.UpdateInfo(mr.Redirect)
		} else {
			break
		}
	}
	n.loadbalancer.Add(n.masterMeta)
	n.loadbalancer.Add(NewNodeMeta(n.selfInfo))
	n.hearbeatMap.Set(n.selfInfo.Id, 0)
	go n.sendHearBeatLoop()
	go func() {
		// 从主节点拉取已存在的节点
		nir, err := n.masterMeta.NodeInfoList()
		if err != nil {
			panic(err)
		}
		for _, v := range nir.NodeInfos {
			if _, ok := n.hearbeatMap.Get(v.Id); !ok {
				n.loadbalancer.Add(NewNodeMeta(v))
				n.hearbeatMap.Set(v.Id, getTime())
			}
		}
	}()
	n.logger.Debug("node become follwer")
}

// Hearbeat 接收心跳
func (n *Node) HearbeatRpc(ctx context.Context, req *nodepb.NodeInfo) (resp *nodepb.HearbeatResp, err error) {
	if n.GetCharacter() != Master {
		return &nodepb.HearbeatResp{Ack: false, Redirect: n.masterMeta.info}, nil
	}
	// 如果当前发送消息的节点是从节点，更新心跳时间。
	_, ok := n.hearbeatMap.Get(req.Id)
	if ok {
		t := getTime()
		n.hearbeatMap.Set(req.Id, t)
		n.logger.Debug("receive hearbeat", zap.Uint64("node id", req.Id))
		return &nodepb.HearbeatResp{Ack: true, Redirect: n.selfInfo}, nil
	}
	return &nodepb.HearbeatResp{Ack: false, Redirect: n.masterMeta.info}, nil
}

// NodeMessage接受消息
func (n *Node) NodeMessageRpc(ctx context.Context, req *nodepb.MessageReq) (resp *nodepb.MessageResp, err error) {
	if n.GetCharacter() != Follower {
		n.logger.Debug("intercrpt a message", zap.Any("message type", req.MessageType), zap.Any("nodeInfo", req.Info))
		return &nodepb.MessageResp{Ack: true}, nil
	}
	// msgType, nodeId, character := DecodeMessage(req.Msg)
	n.logger.Debug("receive a message", zap.Any("message type", req.MessageType), zap.Any("nodeInfo", req.Info))
	if req.MessageType == NodeUp {
		// 节点上线
		n.logger.Debug("node up message:", zap.Uint64("node id", req.Info.Id))
		n.hearbeatMap.Set(req.Info.Id, getTime())
		n.loadbalancer.Add(NewNodeMeta(req.Info))
		if n.nodeUpCallback != nil {
			n.nodeUpCallback(req.Info)
		}
	} else if req.MessageType == NodeDown {
		// 节点下线
		n.logger.Debug("node down message:", zap.Uint64("node id", req.Info.Id))
		nm, _ := n.loadbalancer.Get(req.Info.Id)
		n.loadbalancer.Del(nm)
		n.hearbeatMap.Del(req.Info.Id)
		if n.nodeDownCallback != nil {
			n.nodeDownCallback(req.Info)
		}
	} else if req.MessageType == MasterChange {
		n.logger.Debug("master node change")
		tcp4, mem := DecodeNodeID(req.Info.Id)
		if n.masterMeta != nil {
			n.masterMeta.UpdateId(req.Info.Id)
		} else {
			n.masterMeta = &NodeMeta{
				info: &nodepb.NodeInfo{
					Id:        req.Info.Id,
					Character: Master,
				},
				tcp4addr: string(tcp4[:]),
				memory:   mem,
			}
		}
		// n.masterMeta.UpdateInfo(req.Info)
		nm, _ := n.loadbalancer.Get(req.Info.Id)
		n.loadbalancer.Del(nm)
		n.hearbeatMap.Del(req.Info.Id)

	}
	return &nodepb.MessageResp{Ack: true}, err
}

// Peers获取集群节点
func (n *Node) PeerIdListRpc(ctx context.Context, req *nodepb.EmptyMessage) (resp *nodepb.NodeIdListResp, err error) {
	return &nodepb.NodeIdListResp{NodeIds: n.PeerIds()}, err
}

// Peers获取集群节点
func (n *Node) PeerInfosRpc(ctx context.Context, req *nodepb.EmptyMessage) (resp *nodepb.NodeInfoResp, err error) {
	return &nodepb.NodeInfoResp{NodeInfos: n.PeerInfos()}, err
}

func (n *Node) PeerIds() []uint64 {
	ids := make([]uint64, 0, n.hearbeatMap.Len())
	n.hearbeatMap.ForEach(func(id uint64, _ int64) bool {
		ids = append(ids, id)
		return true
	})
	return ids
}

func (n *Node) PeerInfos() []*nodepb.NodeInfo {
	ids := make([]*nodepb.NodeInfo, 0, n.loadbalancer.Size())
	n.loadbalancer.ForEach(func(_ uint64, nodeMeta *NodeMeta) bool {
		ids = append(ids, nodeMeta.info)
		return true
	})
	return ids
}

func getTime() int64 {
	return time.Now().UnixNano()
}

// ShutdownRpc node down RPC
func (n *Node) LeaveRpc(ctx context.Context, req *nodepb.LeaveReq) (resp *nodepb.LeaveResp, err error) {
	if n.GetCharacter() == Master {
		n.removeNodeMeta(req.Self)
		n.asyncBroadcastMsg(NodeDown, req.Self)
	} else if req.Self.Character == Master {
		// master node leave the cluster,
		// and specfic the current node is new master.
		if req.Master.Id == n.selfInfo.Id {
			// delete the RPC caller's message
			n.hearbeatMap.Del(req.Self.Id)
			nm, _ := n.loadbalancer.Get(req.Self.Id)
			n.loadbalancer.Del(nm)
			n.asyncBroadcastMsg(MasterChange, req.Master)
			n.becomeMaster()

		} else {
			resp, err = n.masterMeta.Leave(req)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}

	} else if req.Self.Character == Follower {
		// if node is not master node ,transpond to master node
		resp, err = n.masterMeta.Leave(req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return &nodepb.LeaveResp{Ack: true}, nil
}

func (n *Node) findTheUsableAndBiggestNodeId(selfId uint64, usable func(uint64) bool) uint64 {
	nodeList := n.PeerIds()
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i] > nodeList[j]
	})
	for i := 0; ; i++ {
		if i == len(nodeList) {
			return 0
		}
		if nodeList[i] != selfId {
			if usable(nodeList[i]) {
				return nodeList[i]
			}
		}
	}
}

func (n *Node) clearNodeMetaAndInfo() {
	ids := make([]uint64, 0, n.hearbeatMap.Len())
	n.hearbeatMap.ForEach(func(id uint64, _ int64) bool {
		ids = append(ids, id)
		return true
	})
	n.hearbeatMap.Del(ids...)
	n.logger.Debug("clear node services meta")
	metas := make([]*NodeMeta, 0, n.loadbalancer.Size())
	n.loadbalancer.ForEach(func(_ uint64, meta *NodeMeta) bool {
		metas = append(metas, meta)
		return true
	})
	n.loadbalancer.Del(metas...)
	n.logger.Debug("clear node hearbeat info")

}

func (n *Node) Leave() error {
	character := n.GetCharacter()
	if character == Follower {
		// call ShutdownRpc to inform master node,it will down
		_, err := n.masterMeta.Leave(&nodepb.LeaveReq{
			Self:   n.selfInfo,
			Master: n.masterMeta.info,
		})
		if err != nil {
			return err
		}
		n.clearNodeMetaAndInfo()
		n.becomeMaster()
		n.masterMeta = nil

	} else if character == Master {
		// 获取所有节点ID,找出除了自己以外最大的ID
		id := n.findTheUsableAndBiggestNodeId(n.selfInfo.Id, func(u uint64) bool {
			meta, ok := n.loadbalancer.Get(u)
			if !ok {
				return false
			}
			resp, err := meta.Leave(&nodepb.LeaveReq{
				Self:   n.selfInfo,
				Master: meta.info,
			})
			if err != nil {
				return false
			}
			if !resp.Ack {
				return false
			}
			n.logger.Debug("master node change:", zap.Any("node meta", meta))
			return true
		})
		if id != 0 {
			// update character to stopp some goroutines about original master node,for clearing origin cluster meta info.
			n.SetCharacter(TemporaryCharacter)
			n.clearNodeMetaAndInfo()
			n.becomeMaster()
			n.masterMeta = nil
		}
	}
	return nil
}

func (n *Node) removeNodeMeta(ni *nodepb.NodeInfo) {
	// 删除主节点信息
	n.hearbeatMap.Del(ni.Id)
	nm, _ := n.loadbalancer.Get(ni.Id)
	n.loadbalancer.Del(nm)
}

// nodeDown
func (n *Node) asyncBroadcastMsg(messageType int64, nodeInfo *nodepb.NodeInfo) {
	n.loadbalancer.ForEach(func(id uint64, meta *NodeMeta) bool {
		if id == n.selfInfo.Id {
			return true
		}
		asyncSendNodeMessage(&nodepb.MessageReq{
			Info:        nodeInfo,
			MessageType: messageType,
		}, meta)
		return true
	})
}

// Meet加入集群,返回ack=true时，表明当前节点应该为主节点，收到响应信息的节点为从节点
// ack=false时，表明当前节点为从节点，收到响应信息的节点参考Redirect字段去寻找主节点
func (n *Node) MeetRpc(ctx context.Context, req *nodepb.NodeInfo) (resp *nodepb.MeetResp, err error) {
	// 当前节点是主节点
	if n.GetCharacter() == Master {
		// 节点上线
		if req.Character == Follower {
			_, exist := n.hearbeatMap.Get(req.Id)
			if exist {
				return &nodepb.MeetResp{Ack: true, Redirect: n.selfInfo}, nil
			}
			// n.logger.Debug(req)
			nodeMeta := NewNodeMeta(req)
			// 广播节点上线的消息
			n.asyncBroadcastMsg(NodeUp, req)
			// 添加节点元数据
			n.hearbeatMap.Set(req.Id, getTime())

			n.loadbalancer.Add(nodeMeta)

			// 返回ack=true
			n.logger.Debug(fmt.Sprintf("new node add cluster id:%d meta:%v", nodeMeta.InstanceID(), nodeMeta))
			return &nodepb.MeetResp{Ack: true, Redirect: n.selfInfo}, err
		}
		// 如果发送心跳的节点也是主节点，
		// 并且发送心跳的节点ID大于当前的节点，启动时间也小于当前节点则当前节点变成从节点
		if n.selfInfo.Id < req.Id && n.selfInfo.StartTime > req.StartTime {
			// 发现新的主节点，转成从节点
			n.SetMasterMeta(req).becomeFollower()
			// n.becomeFollower()
			n.logger.Debug("receive earlier master node meet message,become follwer")
			return &nodepb.MeetResp{Ack: false, Redirect: req}, nil
		}
		n.logger.Debug("receive the other master node meet message,maintain self position.")
		// 反之，告诉发送消息的节点他应该为从节点
		return &nodepb.MeetResp{Ack: true, Redirect: n.selfInfo}, nil
	} else {
		n.logger.Debug(fmt.Sprintf("give a node redirect message %v", n.masterMeta))
		// 当前节点不是主节点，返回重定的主节点向信息
		return &nodepb.MeetResp{Ack: false, Redirect: n.masterMeta.info}, err
	}
}

// reconnect 尝试重连
func (n *Node) reconnect(nodeMeta *NodeMeta, count int) (err error) {
	for i := 0; i < count; i++ {

		hr, err := nodeMeta.Meet(n.selfInfo)
		if err != nil {
			n.logger.Debug("node Meet rpc fail:", zap.Error(err))
			continue
		}
		if !hr.Ack {
			if hr.Redirect.Id != n.selfInfo.Id {
				// 说明那个从节点已经有主节点，去找那个主节点决斗,确定新的主节点
				nm := NewNodeMeta(hr.Redirect)
				mr, err2 := nm.Meet(n.selfInfo)
				if err2 != nil {
					return err
				}
				if mr.Ack {
					n.logger.Debug("find master node while reconnect")
					n.SetMasterMeta(nm.info).becomeFollower()
					return nil
				}
			}
			n.logger.Debug("reconnection successful.....")
			return nil
		} else {
			n.logger.Debug("find master node while reconnect")
			n.SetMasterMeta(hr.Redirect).becomeFollower()
			return nil
		}
	}
	return errors.New("node unresponsive")
}

func (n *Node) nodeHearbeatOutTime(id uint64) {
	// get and remove node meta from loadbalancer,prevent from be repeat used.
	nodeMeta, exist := n.loadbalancer.Get(id)
	if !exist {
		n.logger.Debug("no node meta in loadbalancer...")
		return
	}
	n.loadbalancer.Del(nodeMeta)
	// try to connect...
	err := n.reconnect(nodeMeta, n.retryCount)
	if err != nil {
		n.asyncBroadcastMsg(NodeDown, nodeMeta.info)
		n.logger.Debug("node down....", zap.Any("node", nodeMeta.info))
		return
	}
	// reconnect successful,add node meta to loadbalancer,
	// reset node hearbeat information,
	// and master node will go on checking node hearbeat.
	n.loadbalancer.Add(nodeMeta)
	n.hearbeatMap.Set(id, getTime()+(int64(n.heartbeatTime)>>1))
	n.logger.Debug("node recover hearbeat", zap.Uint64("nodeId", id))
	// if reconnect fail,node info will be delete forever.
}

// checkFollowerHearbeat 检查节点心跳
func (n *Node) checkFollowerHearbeat() {
	if n.GetCharacter() != Master {
		return
	}
	for {
		time.Sleep(time.Duration(n.heartbeatTime))
		if n.GetCharacter() != Master {
			return
		}
		n.logger.Debug("start new once round node heartbeat checking....")
		now := getTime()
		n.hearbeatMap.ForEach(func(id uint64, lt int64) bool {
			if id == n.selfInfo.Id {
				return true
			}
			if now-lt > int64(n.heartbeatTime) {
				// remove node id from hearbeat map at first,
				// prevent repeat scan same node in next time,
				// start retry connection.
				n.hearbeatMap.Del(id)
				n.logger.Debug("node hearbeat out time ,retry connection:", zap.Uint64("node id:", id), zap.Int64("time", now-lt))
				go n.nodeHearbeatOutTime(id)
			}
			return true
		})
	}
}

func (n *Node) sendHearBeatLoop() {
	// 循环向主节点发送心跳
	for {
		time.Sleep(time.Duration(n.heartbeatTime))
		if n.GetCharacter() != Follower {
			return
		}
		n.logger.Debug("hearbeat to master node:", zap.String("address", n.masterMeta.Address()))
		_, err := n.masterMeta.SendHearbeat(n.selfInfo)
		if err != nil {
			n.logger.Debug("hearbeat error:", zap.Error(err))
			// 尝试重连主节点
			err := n.reconnect(n.masterMeta, n.retryCount)
			if err != nil {
				// 重连失败，删除主节点信息，开始寻找新的主节点
				n.loadbalancer.Del(n.masterMeta)
				n.hearbeatMap.Del(n.masterMeta.InstanceID())
				n.logger.Debug("master node down....")
			} else {
				continue
			}
			n.logger.Debug("finding new master.....")
			var peers = n.PeerIds()
			if len(peers) == 0 {
				n.becomeMaster()
				return
			}
			// 从id最大的节点开始寻主节点
			sort.Slice(peers, func(i, j int) bool { return peers[i] > peers[j] })
			n.logger.Debug(fmt.Sprintf("peer ids: %v", peers))
			n.masterMeta.UpdateId(peers[0])
			for i := 0; ; i++ {
				// 如果当前是自己的信息则跳过
				if peers[i] == n.selfInfo.Id {
					n.becomeMaster()
					return
				}
				n.logger.Debug("try to meet node:", zap.Uint64("node id", n.masterMeta.InstanceID()))
				res, err := n.masterMeta.Meet(n.selfInfo)
				if err != nil {
					// 如果所有已知的节点meet请求都失败，自己变成主节点
					if i == len(peers) {
						// 切换成主节点
						n.logger.Debug("no useable node,will switch to master")
						n.becomeMaster()
						return
					}
					n.masterMeta.UpdateId(peers[i])
				} else if !res.Ack {
					// ack=false时说明刚才的消息成功发送，但目标节点不是主节点，
					// 但返回了主节点的地址，下一步就要根据这个地址去寻找主节点
					n.logger.Debug("get redirect master info:", zap.Uint64("node id", n.masterMeta.InstanceID()))
					n.masterMeta.UpdateInfo(res.Redirect)
				} else {
					// 当前节点meet消息成功发送，并且返回ack=true则停止当前循环，
					// 说明找到了主节点，恢复正常心跳。
					n.logger.Debug("found new master node")
					break
				}
			}
		}
	}
}
