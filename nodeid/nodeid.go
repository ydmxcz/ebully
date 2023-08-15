package nodeid

import (
	"encoding/binary"
	"strings"
	"unsafe"
)

const MaxInt64 = 1<<64 - 1

var endianCode = 0x01020304

// 判断当前计算机是大端还是小端，true为小端，false为大端
var isLittleEndian = (*(*byte)(unsafe.Pointer(&endianCode)) == 0x04)

// func GetEndian() bool {
// 	var i int32 = 0x01020304
// 	return (*(*byte)(unsafe.Pointer(&i)) == 0x04)
// }

func unsafeToByte(b string, n int) byte {
	switch n {
	case 1:
		return byte((b[0] - '0') * 1)
	case 2:
		return byte(((b[0] - '0') * 10) + ((b[1] - '0') * 1))
	case 3:
		return byte(((b[0] - '0') * 100) + ((b[1] - '0') * 10) + ((b[2] - '0') * 1))
	default:
		return 0
	}
}

func unsafeToUint16(b string, n int) uint16 {
	if n <= 3 {
		return uint16(unsafeToByte(b, n))
	} else if n == 4 {
		return (uint16(b[0]-'0') * 1000) +
			(uint16(b[1]-'0') * 100) +
			(uint16(b[2]-'0') * 10) +
			(uint16(b[3]-'0') * 1)
	} else if n == 5 {
		// It is possible to overflow in uint16.
		return (uint16(b[0]-'0') * 10000) +
			(uint16(b[1]-'0') * 1000) +
			(uint16(b[2]-'0') * 100) +
			(uint16(b[3]-'0') * 10) +
			(uint16(b[4]-'0') * 1)
	}
	return 0
}

// writeToBytes write a integer number to a slice and return the number of written bytes,
// such as: var num int = 12345 ----(write)----> []byte{'1','2','3','4','5'},
//
//	notes that : while writing,the assess of slice is possible to out range of index.
func writeToBytes(sli []byte, num int) (n int) {
	var stack [24]byte
	var si int
	ti := 0
	si = 0
	if num == 0 {
		sli[0] = '0'
		return 1
	}
	for num > 0 {
		stack[si] = byte(num%10) + '0'
		si++
		num = num / 10
	}
	n = si
	si--
	for si >= 0 {
		sli[ti] = stack[si]
		ti++
		si--
	}
	return n
}

// EncodeNodeID encode an integer of the type uint16 and tcp address string to
// an integer of uint64 as a node id.
// notes that:
// please attention the format of tcp address,there are two format can be right:
// 1) "x.x.x.x:X"
// 2)":X" => "127.0.0.1:X"
func Encode(mem uint16, tcp4 string) uint64 {

	var b [8]byte
	var bi int
	if tcp4 == "" {
		if isLittleEndian {
			binary.LittleEndian.PutUint16(b[:2], mem)
		} else {
			binary.BigEndian.PutUint16(b[6:], mem)
		}
		return *(*uint64)(unsafe.Pointer(&b[0]))
	}
	l, i := 0, strings.Index(tcp4, ":")
	if i == -1 {
		return MaxInt64
	}
	if isLittleEndian {
		if i == 0 { // ip:127.0.0.1
			b[2], b[3], b[4], b[5] = 127, 0, 0, 1
		}
		bi = 2
		for j := 0; j <= i; j++ {
			if tcp4[j] == '.' || j == i {
				if j == l {
					return MaxInt64
				}
				b[bi] = unsafeToByte(tcp4[l:j], j-l)
				bi++
				l = j + 1
			}
		}
		binary.LittleEndian.PutUint16(b[6:], mem)
		binary.LittleEndian.PutUint16(b[:2],
			unsafeToUint16(tcp4[i+1:], len(tcp4[i+1:])))
	} else {
		if i == 0 { // ip:127.0.0.1
			b[2], b[3], b[4], b[5] = 1, 0, 0, 127
		}
		bi = 5
		for j := 0; j <= i; j++ {
			if tcp4[j] == '.' || j == i {
				if j == l {
					return MaxInt64
				}
				b[bi] = unsafeToByte(tcp4[l:j], j-l)
				bi--
				l = j + 1
			}
		}
		binary.BigEndian.PutUint16(b[:2], mem)
		binary.BigEndian.PutUint16(b[6:],
			unsafeToUint16(tcp4[i+1:], len(tcp4[i+1:])))
	}
	// fmt.Println(b)
	return *(*uint64)(unsafe.Pointer(&b[0]))
}

func Weight(id uint64) int {
	b := *(*[8]byte)(unsafe.Pointer(&id))
	if isLittleEndian {
		return int(binary.LittleEndian.Uint16(b[6:]))
	}
	return int(binary.BigEndian.Uint16(b[:2]))
}

func Decode(id uint64) (tcp4 [21]byte, mem int) {
	b := *(*[8]byte)(unsafe.Pointer(&id))
	// fmt.Println(b)
	if isLittleEndian {
		mem = int(binary.LittleEndian.Uint16(b[6:]))
		port := int(binary.LittleEndian.Uint16(b[:2]))
		// fmt.Println(port, mem)
		if port == 0 {
			return
		}
		idx := 0
		for i := 2; i <= 5; i++ {
			idx += writeToBytes(tcp4[idx:], int(b[i]))
			if i == 5 {
				tcp4[idx] = ':'
			} else {
				tcp4[idx] = '.'
			}
			idx++
		}
		writeToBytes(tcp4[idx:], int(port))
	} else {
		mem = int(binary.BigEndian.Uint16(b[:2]))
		port := int(binary.BigEndian.Uint16(b[6:]))
		if port == 0 {
			return
		}
		idx := 0
		for i := 5; i >= 2; i-- {
			idx += writeToBytes(tcp4[idx:], int(b[i]))
			if i == 5 {
				tcp4[idx] = ':'
			} else {
				tcp4[idx] = '.'
			}
			idx++
		}
		writeToBytes(tcp4[idx:], int(port))
	}
	return
}
