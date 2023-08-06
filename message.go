package ebully

import "encoding/binary"

func DecodeMessage(msg []byte) (msgType uint16,
	nodeId uint64, character CharacterType) {

	if len(msg) < 12 {
		return
	}
	return binary.LittleEndian.Uint16(msg[0:2]),
		binary.LittleEndian.Uint64(msg[2:10]),
		CharacterType(msg[10])
}

func EncodeMessage(msgType uint16, nodeId uint64,
	character CharacterType) [12]byte {

	var msg [12]byte
	binary.LittleEndian.PutUint16(msg[0:2], msgType)
	binary.LittleEndian.PutUint64(msg[2:10], nodeId)
	msg[10] = byte(character)
	msg[11] = messageSplitSymbol
	return msg
}
