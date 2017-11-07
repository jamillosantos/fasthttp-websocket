package websocket

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

// ConnectionState represents the state of the websocket connection.
type ConnectionState byte

const (
	// ConnectionStateConnecting represents the initial state of a connection
	ConnectionStateConnecting = iota
	// ConnectionStateOpen represents a opened health connection
	ConnectionStateOpen
	// ConnectionStateClosing represents a connection that is closing
	ConnectionStateClosing
	// ConnectionStateClosed represents a closed connection
	ConnectionStateClosed
)

// ConnectionCloseReason represents the reason informed by the endpoint for
// closing the connection.
type ConnectionCloseReason uint16

const (
	// ConnectionCloseReasonNormal happens when the endpoint simply want to end
	// the connection
	ConnectionCloseReasonNormal ConnectionCloseReason = 1000
	// ConnectionCloseReasonGoingDown happens when the server is going down, or
	// the client is navigating away to another page.
	ConnectionCloseReasonGoingDown ConnectionCloseReason = 1001
	// ConnectionCloseReasonProtocolError happens when there is any error on the
	// protocol
	ConnectionCloseReasonProtocolError ConnectionCloseReason = 1002
	// ConnectionCloseReasonDataTypeUnsupported happens when the endpoint
	// receives a datatype it cannot accept.
	ConnectionCloseReasonDataTypeUnsupported ConnectionCloseReason = 1003
	// ConnectionCloseReasonInconsistentType happens when the endpoint receives
	// a message inconsistent with the type of the message
	ConnectionCloseReasonInconsistentType ConnectionCloseReason = 1007
	// ConnectionCloseReasonPolicyViolation happens when the endpoint receives a
	// message that violates its policy
	ConnectionCloseReasonPolicyViolation ConnectionCloseReason = 1008
	// ConnectionCloseReasonMessageTooBig happens when the endpoint receives a
	// message bigger than it can process.
	ConnectionCloseReasonMessageTooBig ConnectionCloseReason = 1009
	// ConnectionCloseReasonCouldNotNegotiateExtensions happens when the client
	// and the server fail to negotiate the extensions.
	ConnectionCloseReasonCouldNotNegotiateExtensions ConnectionCloseReason = 1010
	// ConnectionCloseReasonUnexpected happens when the server is terminating
	// the connection because it encoutered an unexpected condition
	ConnectionCloseReasonUnexpected ConnectionCloseReason = 1011
)

// MessageType represents the type of message defined by the RFC 6455
type MessageType byte

const (
	// MessageTypeContinuation represents a continuation
	MessageTypeContinuation MessageType = 0
	// MessageTypeText represents a text frame
	MessageTypeText MessageType = 1
	// MessageTypeBinary represents a binary frame
	MessageTypeBinary MessageType = 2
	// MessageTypeConnectionClose represents a closing message and the
	// connection will be closed right away
	MessageTypeConnectionClose MessageType = 8
	// MessageTypePing represents a ping frame
	MessageTypePing MessageType = 9
	// MessageTypePong represents a pong frame
	MessageTypePong MessageType = 10
)

var (
	errorMissingMaskingKey = errors.New("protocol error: missing masking key")
)

// Connection is the minimum representation of a websocket connection
type Connection interface {
	Init(context *ConnectionContext)
	Reset()

	Conn() net.Conn

	State() ConnectionState

	Context() interface{}
	SetContext(value interface{})

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
	context    interface{}
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

// Init implements the websocket.Connection.Init
func (c *BaseConnection) Init(ctx *ConnectionContext) {
	c.compressed = ctx.Compressed
	c.conn = ctx.Conn
	c.state = ConnectionStateOpen
}

// Conn implements the websocket.Connection.Conn
func (c *BaseConnection) Conn() net.Conn {
	return c.conn
}

// State implements the websocket.Connection.State
func (c *BaseConnection) State() ConnectionState {
	return c.state
}

func (c *BaseConnection) Context() interface{} {
	return c.context
}

func (c *BaseConnection) SetContext(value interface{}) {
	c.context = value
}

// Read implements the websocket.Connection.Read
func (c *BaseConnection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

// ReadPacket implements the websocket.Connection.ReadPacket
func (c *BaseConnection) ReadPacket() (byte, []byte, error) {
	n, err := c.Read(c.readBuff)
	if n == 0 {
		// No data
		return 0, nil, nil
	}
	if (err != nil) && (err != io.EOF) {
		return 0, nil, err
	}
	_, _, _, _, opcode, _, maskingKey, payload, err := DecodePacket(c.readBuff[:n])

	if err != nil {
		return 0, nil, err
	}

	if maskingKey == nil {
		err = c.CloseWithReason(ConnectionCloseReasonProtocolError)
		if err != nil {
			return 0, nil, err
		}
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

// ReadPacketTimeout implements the websocket.Connection.ReadPacketTimeout
func (c *BaseConnection) ReadPacketTimeout(timeout time.Duration) (byte, []byte, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err == nil {
		return 0, nil, err
	}
	return c.ReadPacket()
}

// Write implements the websocket.Connection.Write
func (c *BaseConnection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *BaseConnection) preparePacket(opcode byte, payload []byte) ([]byte, error) {
	return EncodePacket(true, c.compressed, false, false, opcode, uint64(len(payload)), nil, payload)
}

// WritePacket implements the websocket.Connection.WritePacket
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

// WritePacketTimeout implements the websocket.Connection.WritePacketTimeout
func (c *BaseConnection) WritePacketTimeout(timeout time.Duration, opcode byte, data []byte) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	return c.WritePacket(opcode, data)
}

// IsClosed implements the websocket.Connection.IsClosed
func (c *BaseConnection) IsClosed() bool {
	return c.state == ConnectionStateClosed
}

// Close implements the websocket.Connection.Close
func (c *BaseConnection) Close() error {
	return c.CloseWithReason(ConnectionCloseReasonNormal)
}

// CloseWithReason implements the websocket.Connection.CloseWithReason
func (c *BaseConnection) CloseWithReason(reason ConnectionCloseReason) error {
	c.state = ConnectionStateClosing
	var payload [2]byte
	binary.BigEndian.PutUint16(payload[:], uint16(reason))
	return c.WritePacket(OPCodeConnectionCloseFrame, payload[:])
}

// Terminate implements the websocket.Connection.Terminate
func (c *BaseConnection) Terminate() error {
	err := c.conn.Close()
	if err == nil {
		c.state = ConnectionStateClosed
	}
	return err
}
