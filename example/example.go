package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ydmxcz/ebully"
	"go.uber.org/zap"
)

// note that:the current way of running one node depend on command line parameters.
// compile command:
// [go build .] or [go run example.go [-args]...]
// given five command here to run a node ,You can use one or more of these or
// write orther correct startup commands to start any number of nodes on different terminals:)
// on general,using three nodes is most convenient.
// ./test -selfAddr="192.168.0.200:8080" -nodeWeight=16 -httpServiceAddr="192.168.0.200:8081" -tcpReverseProxyAddr="192.168.0.200:8082"
// ./test -selfAddr="192.168.0.200:8180" -masterAddr="192.168.0.200:8080" -nodeWeight=14 -httpServiceAddr="192.168.0.200:8181" -tcpReverseProxyAddr="192.168.0.200:8182"
// ./test -selfAddr="192.168.0.200:8280" -masterAddr="192.168.0.200:8080" -nodeWeight=12 -httpServiceAddr="192.168.0.200:8281" -tcpReverseProxyAddr="192.168.0.200:8282"
// ./test -selfAddr="192.168.0.200:8380" -masterAddr="192.168.0.200:8080" -nodeWeight=10 -httpServiceAddr="192.168.0.200:8381" -tcpReverseProxyAddr="192.168.0.200:8382"
// ./test -selfAddr="192.168.0.200:8480" -masterAddr="192.168.0.200:8080" -nodeWeight=8 -httpServiceAddr="192.168.0.200:8481" -tcpReverseProxyAddr="192.168.0.200:8482"

func main() {
	selfAddr := flag.String("selfAddr", "", "self address")
	nodeWeight := flag.Int("nodeWeight", 8, "node weight")
	masterAddr := flag.String("masterAddr", "", "master address")
	serviceAddr := flag.String("httpServiceAddr", "", "http service addr")
	tcpReverseProxyAddr := flag.String("tcpReverseProxyAddr", "", "tcp reverse proxy addr")
	flag.Parse()

	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zap.DebugLevel))
	if err != nil {
		panic(err)
	}
	cfg := ebully.Config{
		ID:             ebully.EncodeNodeID(uint16(*nodeWeight), *selfAddr),
		RetryCount:     3,
		SelfAddr:       *selfAddr,
		MasterAddr:     *masterAddr,
		RpcTimeOut:     time.Duration(time.Second),
		Logger:         logger,
		ServiceAddress: *serviceAddr,
		HeartBeatTime:  5 * time.Second,
		EnableTcpProxy: true,
		Listener: func() net.Listener {
			l, err := net.Listen("tcp", *tcpReverseProxyAddr)
			if err != nil {
				panic(err)
			}
			return l
		}(),
	}
	logger.Info("finish load config", zap.Any("config", cfg))
	node := ebully.NewEBully(cfg)
	node.Start()
	http.HandleFunc("/peers", func(resp http.ResponseWriter, _ *http.Request) {
		peerInfos := node.PeerList()
		b, err := json.Marshal(peerInfos)
		if err != nil {
			resp.Write([]byte(err.Error()))
		}
		resp.Write(b)
	})
	http.HandleFunc("/hello", func(resp http.ResponseWriter, _ *http.Request) {
		resp.Write([]byte(fmt.Sprintf("hello, here is %s,node id: %d", *serviceAddr, node.ID())))
	})

	http.HandleFunc("/meet", func(resp http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		id := query.Get("master")
		t, err := net.ResolveTCPAddr("tcp4", id)
		if err != nil {
			resp.Write([]byte(err.Error()))
			return
		}
		err = node.Meet(t.String())
		if err != nil {
			resp.Write([]byte(err.Error()))
			return
		}

		resp.Write([]byte(fmt.Sprintf("node %s meet cluster success on %v !", *serviceAddr, time.Now())))
	})

	http.HandleFunc("/leave", func(resp http.ResponseWriter, _ *http.Request) {
		err := node.Leave()
		if err != nil {
			resp.Write([]byte(err.Error()))
			return
		}
		resp.Write([]byte(fmt.Sprintf("node %s leave cluster success on %v !", *serviceAddr, time.Now())))
	})
	logger.Info("http server is running...")
	err = http.ListenAndServe(*serviceAddr, nil)
	if err != nil {
		panic(err)
	}
}
