package websocket

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"sync"
)

const (
	bit1 byte = 1 << 7
	bit2 byte = 1 << 6
	bit3 byte = 1 << 5
	bit4 byte = 1 << 4
	bit5 byte = 1 << 3
	bit6 byte = 1 << 2
	bit7 byte = 1 << 1
	bit8 byte = 1
)

const (
	positionFinRsvsOpCode                      byte = 0
	positionMaskPayloadLen                     byte = 1
	positionMaskPayloadLenExtended             byte = 2
	positionMaskPayloadLenExtended16bitsEnding byte = positionMaskPayloadLenExtended + 2
	positionMaskPayloadLenExtended64bitsEnding byte = positionMaskPayloadLenExtended + 8
)

const (
	maskFin        byte = bit1
	maskRsv1       byte = bit2
	maskRsv2       byte = bit3
	maskRsv3       byte = bit4
	maskOpCode     byte = bit5 | bit6 | bit7 | bit8
	maskMask       byte = bit1
	maskPayloadLen byte = ^maskMask
)

const (
	payloadLen16bits = 126
	payloadLen64bits = 127
)

const (
	// OPCodeContinuationFrame is the continuation operation code of the dataframe
	OPCodeContinuationFrame byte = 0
	// OPCodeTextFrame is the text message operation code of the dataframe
	OPCodeTextFrame byte = 1
	// OPCodeBinaryFrame is the binary message operation code of the dataframe
	OPCodeBinaryFrame byte = 2
	// OPCodeConnectionCloseFrame is the connection close operation code of the dataframe
	OPCodeConnectionCloseFrame byte = 8
	// OPCodePingFrame is the ping operation code of the dataframe
	OPCodePingFrame byte = 9
	// OPCodePongFrame is the pong operation code of the dataframe
	OPCodePongFrame byte = 10
)

// DecodePacket splits all the information from the raw packet and return it.
func DecodePacket(buff []byte) (fin bool, rsv1 bool, rsv2 bool, rsv3 bool, opcode byte, payloadLen uint64, maskingKey []byte, payload []byte, err error) {
	// TODO IMPORTANT! Check for acessing out of bound indexes.

	// 1st byte
	fin = ((buff[positionFinRsvsOpCode] & maskFin) == maskFin)
	rsv1 = ((buff[positionFinRsvsOpCode] & maskRsv1) == maskFin)
	rsv2 = ((buff[positionFinRsvsOpCode] & maskRsv2) == maskFin)
	rsv3 = ((buff[positionFinRsvsOpCode] & maskRsv3) == maskFin)
	opcode = buff[positionFinRsvsOpCode] & maskOpCode

	// 2nd byte
	masked := (buff[positionMaskPayloadLen] & maskMask) == maskMask
	payloadLen = uint64(buff[positionMaskPayloadLen] & (maskPayloadLen))

	// Check if the payload length is extended
	startAt := positionMaskPayloadLenExtended
	if payloadLen == payloadLen16bits {
		startAt += 2
		payloadLen = uint64(binary.BigEndian.Uint16(buff[positionMaskPayloadLenExtended:positionMaskPayloadLenExtended16bitsEnding]))
	} else if payloadLen == payloadLen64bits {
		startAt += 8
		payloadLen = binary.BigEndian.Uint64(buff[positionMaskPayloadLenExtended:positionMaskPayloadLenExtended64bitsEnding])
	}

	// Check the masking key
	if masked {
		maskingKey = buff[2:6]
		startAt += 4
	}
	payload = buff[startAt:(uint64(startAt) + payloadLen)]
	return
}

// Unmask unmasks a masked payload.
func Unmask(buff, mask []byte) {
	l := uint64(len(buff))
	masklen := uint64(len(mask))
	// Unmask the payload
	for i := uint64(0); i < l; i++ {
		buff[i] = buff[i] ^ mask[i%masklen]
	}
}

const deflateBufferDefaultSize = 1024

var deflateBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, deflateBufferDefaultSize)
	},
}

// Deflate deflates a flated package
func Deflate(dst, src []byte) ([]byte, error) {
	buff := deflateBufferPool.Get().([]byte)
	defer deflateBufferPool.Put(buff)

	reader := flate.NewReader(bytes.NewReader(append(src, []byte{0x00, 0x00, 0xff, 0xff}...)))
	n, err := reader.Read(buff)

	for (err == nil) && (n > 0) {
		dst = append(dst, buff[:n]...)
		if n < len(buff) {
			break
		}
		n, err = reader.Read(buff)
	}
	if err == nil {
		return dst, err
	}
	return nil, err
}
