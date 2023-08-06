package ebully_test

import (
	"fmt"
	"testing"

	"github.com/ydmxcz/ebully"
	nodepb "github.com/ydmxcz/ebully/node"
)

type testNode struct {
	*nodepb.NodeInfo
	Name string
}

func TestSelectMaster(t *testing.T) {

	nodes := []testNode{
		{
			NodeInfo: &nodepb.NodeInfo{
				Id:        ebully.EncodeNodeID(16, "192.168.0.1:8080"),
				StartTime: 10000,
				Character: ebully.Master,
			},
			Name: "Node-1",
		},
		{
			NodeInfo: &nodepb.NodeInfo{
				Id:        ebully.EncodeNodeID(8, "192.168.0.1:8081"),
				StartTime: 11000,
				Character: ebully.Follower,
			},
			Name: "Node-2",
		}, {
			NodeInfo: &nodepb.NodeInfo{
				Id:        ebully.EncodeNodeID(4, "192.168.0.1:8082"),
				StartTime: 9000,
				Character: ebully.Follower,
			},
			Name: "Node-3",
		},
	}
	for _, n := range nodes {
		fmt.Println("node-name:", n.Name, " node-id:", n.Id)
	}
}

func findTheUsableAndBiggestNodeId(selfId int64, nodeList []int64, verify func(int64) bool) int64 {
	i := 0
	for {
		if i == len(nodeList) {
			return -1
		}
		if nodeList[i] != selfId {
			if verify(nodeList[i]) {
				return nodeList[i]
			} else {
				i++
			}
		} else {
			i++
		}
	}
}

func TestRand(t *testing.T) {
	arr := []int64{9, 8, 7, 6, 5, 4, 3, 2, 1}
	res := findTheUsableAndBiggestNodeId(5, arr, func(i int64) bool {
		return i < 3
	})
	fmt.Println(res)
	fmt.Println(ebully.EncodeNodeID(0, "192.168.0.200:8081"))
	tcp4, mem := ebully.DecodeNodeID(4723504784088976)
	fmt.Println(string(tcp4[:]), mem)
}
