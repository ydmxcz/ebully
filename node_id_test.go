package ebully_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ydmxcz/ebully"
)

func TestNodeID(t *testing.T) {
	addr := "192.168.0.192:6666"
	id := ebully.EncodeNodeID(0, addr)
	fmt.Println(id)
	tcp, mem := ebully.DecodeNodeID(id)
	fmt.Println(id, string(tcp[:]), mem)

}

func vaildNodeId(mem uint16, tcp4 string) {
	id := ebully.EncodeNodeID(mem, tcp4)
	tcp, m := ebully.DecodeNodeID(id)
	fmt.Println(id, string(tcp[:]), m)

}

func TestGenNodeId(t *testing.T) {
	vaildNodeId(16, "192.168.0.191:80")
	vaildNodeId(8, "192.168.0.191:80")
	vaildNodeId(4, "192.168.0.191:80")
}

func TestTime(t *testing.T) {
	t1 := time.Now().UnixNano()
	time.Sleep(time.Second * 5)
	t2 := time.Now().UnixNano()
	fmt.Println(time.Second*5, t2-t1)
}
// 4713609179430992
// 2461809365745744
// 1335909458903120