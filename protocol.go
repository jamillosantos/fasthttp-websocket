package websocket

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"io"
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
	payloadLen16bits       = byte(126)
	payloadLen64bits       = byte(127)
	payloadLen16bitsUint64 = uint64(126)
	payloadLen64bitsUint64 = uint64(127)
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

var (
	ErrUnexpectedEndOfPacket = errors.New("Unexpected end of packet")
	errorWrongMaskKey        = errors.New("Wrong mask key")
)

// IsUnexpectedEndOfPacket checks if the given error is of type unexpected end of packet
func IsUnexpectedEndOfPacket(err error) bool {
	return err == ErrUnexpectedEndOfPacket
}

// DecodePacket splits all the information from the raw packet and return it.
func DecodePacket(buff []byte) (fin bool, rsv1 bool, rsv2 bool, rsv3 bool, opcode byte, payloadLen uint64, maskingKey []byte, payload []byte, err error) {
	buffLen := uint64(len(buff))
	if buffLen < 2 {
		return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
	}

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
	if payloadLen == payloadLen16bitsUint64 {
		startAt += 2
		if buffLen < uint64(positionMaskPayloadLenExtended16bitsEnding) {
			return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
		}
		payloadLen = uint64(binary.BigEndian.Uint16(buff[positionMaskPayloadLenExtended:positionMaskPayloadLenExtended16bitsEnding]))
	} else if payloadLen == payloadLen64bitsUint64 {
		startAt += 8
		if buffLen < uint64(startAt) {
			return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
		}
		payloadLen = binary.BigEndian.Uint64(buff[positionMaskPayloadLenExtended:positionMaskPayloadLenExtended64bitsEnding])
	} else if (payloadLen + uint64(positionMaskPayloadLen)) > buffLen {
		return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
	}

	// Check the masking key
	if masked {
		startAt += 4
		if buffLen < uint64(startAt) {
			return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
		}
		maskingKey = buff[(startAt - 4):startAt]
	}
	if buffLen < uint64(startAt)+payloadLen {
		return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
	}
	payload = buff[startAt:(uint64(startAt) + payloadLen)]
	return
}

func DecodePacketFromReader(reader io.Reader, buff []byte) (fin bool, rsv1 bool, rsv2 bool, rsv3 bool, opcode byte, payloadLen uint64, maskingKey []byte, payload []byte, err error) {
	var n int
	buffLen := uint64(len(buff))
	n, err = reader.Read(buff[:2])
	if (err != nil && err != io.EOF) || (n < 2) {
		return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
	}

	// 1st byte
	fin = (buff[positionFinRsvsOpCode] & maskFin) == maskFin
	rsv1 = (buff[positionFinRsvsOpCode] & maskRsv1) == maskFin
	rsv2 = (buff[positionFinRsvsOpCode] & maskRsv2) == maskFin
	rsv3 = (buff[positionFinRsvsOpCode] & maskRsv3) == maskFin
	opcode = buff[positionFinRsvsOpCode] & maskOpCode

	// 2nd byte
	masked := (buff[positionMaskPayloadLen] & maskMask) == maskMask
	pl := buff[positionMaskPayloadLen] & (maskPayloadLen)
	payloadLen = uint64(pl)

	// Check if the payload length is extended
	startAt := positionMaskPayloadLenExtended
	if payloadLen == payloadLen16bitsUint64 {
		n, err = reader.Read(buff[:2])
		if (n != 2) || (err != nil) {
			err = ErrUnexpectedEndOfPacket
			return
		}
		payloadLen = uint64(binary.BigEndian.Uint16(buff[:2]))
	} else if payloadLen == payloadLen64bitsUint64 {
		n, err = reader.Read(buff[:8])
		if (n != 8) || (err != nil) {
			err = ErrUnexpectedEndOfPacket
			return
		}
		payloadLen = binary.BigEndian.Uint64(buff[:8])
	}

	// Check the masking key
	if masked {
		n, err = reader.Read(buff[:4])
		if buffLen < uint64(startAt) {
			return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
		}
		maskingKey = make([]byte, 4)
		copy(maskingKey, buff[:4])
	}
	if buffLen < payloadLen {
		payload = make([]byte, payloadLen)
		n, err = reader.Read(payload)
	} else {
		n, err = reader.Read(buff[:payloadLen])
		if (err != nil && err != io.EOF) || uint64(n) < payloadLen {
			return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
		}
		if err != nil && err != io.EOF {
			return false, false, false, false, 0, 0, nil, nil, ErrUnexpectedEndOfPacket
		}
		payload = buff[:n]
	}
	if err != nil && err != io.EOF {
		return false, false, false, false, 0, 0, nil, nil, err
	}
	err = nil
	return
}

// Unmask unmasks a masked payload.
//
// There is no Mask method. Since the masking procedure is a bitwise not,
// applying Unmask will toggle the un/masking.
func Unmask(buff, mask []byte) {
	l := uint64(len(buff))
	masklen := uint64(len(mask))
	// Unmask the payload
	for i := uint64(0); i < l; i++ {
		buff[i] = buff[i] ^ mask[i%masklen]
	}
}

// EncodePacket generates a byte array with the packet encoded according with
// the RFC 6455
func EncodePacket(fin bool, rsv1 bool, rsv2 bool, rsv3 bool, opcode byte, payloadLen uint64, maskingKey []byte, payload []byte) ([]byte, error) {
	if (maskingKey != nil) && (len(maskingKey) != 4) {
		return nil, errorWrongMaskKey
	}

	l := 2 // Default header
	if payloadLen > math.MaxUint16 {
		l += 8 // + 64-bit length
	} else if payloadLen >= uint64(payloadLen16bits) {
		l += 2 // + 16-bit length
	}
	if maskingKey != nil {
		l += 4 // + masking key
	}

	dst := make([]byte, l)
	if fin {
		dst[positionFinRsvsOpCode] = dst[positionFinRsvsOpCode] | maskFin
	}
	if rsv1 {
		dst[positionFinRsvsOpCode] = dst[positionFinRsvsOpCode] | maskRsv1
	}
	if rsv2 {
		dst[positionFinRsvsOpCode] = dst[positionFinRsvsOpCode] | maskRsv2
	}
	if rsv3 {
		dst[positionFinRsvsOpCode] = dst[positionFinRsvsOpCode] | maskRsv3
	}
	dst[positionFinRsvsOpCode] = dst[positionFinRsvsOpCode] | (opcode & maskOpCode)

	maskingValue := byte(0)
	if maskingKey != nil {
		maskingValue = maskMask
	}
	startAt := positionMaskPayloadLenExtended
	if payloadLen < uint64(payloadLen16bits) {
		dst[positionMaskPayloadLen] = maskingValue | (maskPayloadLen & byte(payloadLen))
	} else if payloadLen <= math.MaxUint16 {
		dst[positionMaskPayloadLen] = maskingValue | payloadLen16bits
		binary.BigEndian.PutUint16(dst[positionMaskPayloadLenExtended:], uint16(payloadLen))
		startAt += 2
	} else {
		dst[positionMaskPayloadLen] = maskingValue | payloadLen64bits
		binary.BigEndian.PutUint64(dst[positionMaskPayloadLenExtended:], payloadLen)
		startAt += 8
	}

	if maskingKey != nil {
		copy(dst[startAt:], maskingKey)
	}

	dst = append(dst, payload...)
	return dst, nil
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

// Flate compress the given src into the dst buffer
// It returns the amount of bytes written or an error.
func Flate(dst, src []byte) ([]byte, int, error) {
	b := bytes.NewBuffer(dst)
	writer, err := flate.NewWriter(b, flate.BestCompression)
	if err != nil {
		return nil, 0, err
	}
	_, err = writer.Write(src)
	if err != nil {
		return nil, 0, err
	}
	err = writer.Flush()
	if err != nil {
		return nil, 0, err
	}
	l := b.Len() - 4
	return b.Bytes()[:l], l, nil
}
