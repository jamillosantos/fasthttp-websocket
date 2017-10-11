package websocket

import (
	"net"
	"io"
	"time"
	"encoding/binary"
	"errors"
)

type ConnectionState byte

const (
	ConnectionStateConnecting = iota
	ConnectionStateOpen
	ConnectionStateClosing
	ConnectionStateClosed
)

type ConnectionCloseReason uint16

const (
	ConnectionCloseReasonNormal                      ConnectionCloseReason = 1000
	ConnectionCloseReasonGoingDown                   ConnectionCloseReason = 1001
	ConnectionCloseReasonProtocolError               ConnectionCloseReason = 1002
	ConnectionCloseReasonDataTypeUnsupported         ConnectionCloseReason = 1003
	ConnectionCloseReasonInconsistentType            ConnectionCloseReason = 1007
	ConnectionCloseReasonPolicyViolation             ConnectionCloseReason = 1008
	ConnectionCloseReasonMessageTooBig               ConnectionCloseReason = 1009
	ConnectionCloseReasonCouldNotNegotiateExtensions ConnectionCloseReason = 1010
	ConnectionCloseReasonUnexpected                  ConnectionCloseReason = 1011
)

type MessageType byte

const (
	MessageTypeContinuation    MessageType = 0
	MessageTypeText            MessageType = 1
	MessageTypeBinary          MessageType = 2
	MessageTypeConnectionClose MessageType = 8
	MessageTypePing            MessageType = 9
	MessageTypePong            MessageType = 10
)

var (
	errorMissingMaskingKey = errors.New("Protocol error: Missing masking key.")
)

type Connection interface {
	Init(context *ConnectionContext)

	Conn() net.Conn

	State() ConnectionState

	Read(buffer []byte) (int, error)
	Write(data []byte) (int, error)

	ReadPacket() (byte, []byte, error)
	ReadPacketTimeout(timeout time.Duration) (byte, []byte, error)

	WritePacket(opcode byte, payload []byte) error
	WritePacketTimeout(timeout time.Duration, opcode byte, payload []byte) error

	ReadMessage() (MessageType, []byte, error)
	ReadMessageTimeout(timeout time.Duration) (MessageType, []byte, error)
	WriteMessage(opcode MessageType, payload []byte) error
	WriteMessageTimeout(timeout time.Duration, opcode MessageType, payload []byte) error

	IsClosed() bool
	Close() error
	CloseWithReason(reason ConnectionCloseReason) error
	Terminate() error
}

// BaseConnection represents a connection with a client
type BaseConnection struct {
	readBuff   []byte
	conn       net.Conn
	state      ConnectionState
	compressed bool
}

// NewConn initialized and return a new websocket.BaseConnection instance
func NewConn(conn net.Conn) *BaseConnection {
	return &BaseConnection{
		readBuff: make([]byte, 1024*8),
		conn:     conn,
	}
}

// Reset cleans up all the data and prepare the instance for being placed back
// on the pool, for avoiding allocation.
func (c *BaseConnection) Reset() {
	c.conn = nil
	c.compressed = false
	c.state = ConnectionStateClosed
}

func (c *BaseConnection) Init(ctx *ConnectionContext) {
	c.compressed = ctx.Compressed
	c.conn = ctx.Conn
	c.state = ConnectionStateOpen
}

func (c *BaseConnection) Conn() net.Conn {
	return c.conn
}

func (c *BaseConnection) State() ConnectionState {
	return c.state
}

func (c *BaseConnection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *BaseConnection) ReadPacket() (byte, []byte, error) {
	n, err := c.Read(c.readBuff)
	if n == 0 {
		// No data
		return 0, nil, nil
	}
	if err == io.EOF {
		err = nil
	}
	_, _, _, _, opcode, _, maskingKey, payload, err := DecodePacket(c.readBuff[:n])
	if maskingKey == nil {
		c.CloseWithReason(ConnectionCloseReasonProtocolError)
		return 0, nil, errorMissingMaskingKey
	}
	Unmask(payload, maskingKey)
	if c.compressed && (opcode != OPCodeConnectionCloseFrame) {
		dpayload, err := Deflate(make([]byte, 0, 1024), payload)
		if err != nil {
			return 0, nil, err
		}
		return opcode, dpayload, nil
	}
	return opcode, payload, nil
}

func (c *BaseConnection) ReadPacketTimeout(timeout time.Duration) (byte, []byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	return c.ReadPacket()
}

func (c *BaseConnection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *BaseConnection) preparePacket(opcode byte, payload []byte) ([]byte, error) {
	return EncodePacket(true, c.compressed, false, false, opcode, uint64(len(payload)), nil, payload)
}

func (c *BaseConnection) WritePacket(opcode byte, data []byte) error {
	var err error
	if c.compressed {
		data, _, err = Flate(make([]byte, 0, 1024), data)
		if err != nil {
			return err
		}
	}
	packet, err := c.preparePacket(opcode, data)
	if err != nil {
		return err
	}
	_, err = c.Write(packet)
	return err
}

func (c *BaseConnection) WritePacketTimeout(timeout time.Duration, opcode byte, data []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(timeout))
	return c.WritePacket(opcode, data)
}

func (c *BaseConnection) IsClosed() bool {
	return c.state == ConnectionStateClosed
}

func (c *BaseConnection) Close() error {
	return c.CloseWithReason(ConnectionCloseReasonNormal)
}

func (c *BaseConnection) CloseWithReason(reason ConnectionCloseReason) error {
	c.state = ConnectionStateClosing
	var payload [2]byte
	binary.BigEndian.PutUint16(payload[:], uint16(reason))
	return c.WritePacket(OPCodeConnectionCloseFrame, payload[:])
}

func (c *BaseConnection) Terminate() error {
	err := c.conn.Close()
	if err == nil {
		c.state = ConnectionStateClosed
	}
	return err
}
