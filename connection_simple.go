package websocket

import (
	"net"
	"time"
	"unicode/utf8"
	"encoding/binary"
	"golang.org/x/text/encoding"
	"log"
)

// SimpleConnection represents a connection with a client
type SimpleConnection struct {
	BaseConnection
	lastMessageAt time.Time
}

// NewSimpleConn initialized and return a new websocket.BaseConnection instance
func NewSimpleConn(conn net.Conn) *SimpleConnection {
	return &SimpleConnection{
		BaseConnection: *NewConn(conn),
	}
}

// ReadMessage implements the websocket.Connection.ReadMessage method
func (c *SimpleConnection) ReadMessage() (MessageType, []byte, error) {
	if c.state == ConnectionStateClosing {
		return 0, nil, ErrConnectionClosing
	}
	if c.state == ConnectionStateClosing {
		return 0, nil, ErrConnectionClosed
	}

	var (
		npayload []byte
		nopcode  MessageType
	)
	for {
		fin, opc, payload, err := c.ReadPacket()

		if err != nil {
			return 0, nil, err
		}

		if payload == nil {
			return 0, nil, nil
		}

		opcode := MessageType(opc)
		c.lastMessageAt = time.Now()
		switch opcode {
		case MessageTypePing, MessageTypePong, MessageTypeConnectionClose: // Control frames
			if len(payload) > 125 { // If control frames payload bigger than 125 (could not find on the RFC. However, Autobahn Testsuite implements this way)
				c.CloseWithReason(ConnectionCloseReasonProtocolError)
				c.Terminate()
				return 0, nil, nil
			}
			switch opcode {
			case MessageTypePing:
				if c.state == ConnectionStateOpen {
					// Respond the ping message with payload
					err = c.WritePacketTimeout(time.Millisecond*10, OPCodePongFrame, payload)
					if err != nil {
						return 0, nil, err
					}
				}
				if npayload == nil {
					return 0, nil, nil
				}
			case MessageTypePong:
				if npayload != nil {
					return 0, nil, nil
				}
			case MessageTypeConnectionClose:
				if len(payload) < 2 && len(payload) != 0 {
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, nil
				}
				closingReason := ConnectionCloseReasonNormal
				if len(payload) >= 2 {
					closingReason = ConnectionCloseReason(uint16(binary.BigEndian.Uint16(payload[:2])))
					payload = payload[2:]
				}

				switch closingReason {
				case ConnectionCloseReasonNormal, ConnectionCloseReasonGoingDown, ConnectionCloseReasonProtocolError,  ConnectionCloseReasonDataTypeUnsupported, ConnectionCloseReasonInconsistentType, ConnectionCloseReasonPolicyViolation, ConnectionCloseReasonMessageTooBig, ConnectionCloseReasonCouldNotNegotiateExtensions, ConnectionCloseReasonUnexpected:
				default:
					if closingReason < 3000 || closingReason >= 5000 {
						c.CloseWithReason(ConnectionCloseReasonProtocolError)
						c.Terminate()
						return 0, nil, ErrWrongClosingCode
					}
				}

				if !utf8.Valid(payload) {
					log.Println(string(payload))
					c.CloseWithReason(ConnectionCloseReasonInconsistentType)
					c.Terminate()
					return 0, nil, encoding.ErrInvalidUTF8
				}
				c.state = ConnectionStateClosing
				c.Close()
				err = c.Terminate()
				return 0, nil, err
			}
		case MessageTypeContinuation, MessageTypeBinary, MessageTypeText:
			if fin {
				if opcode == MessageTypeContinuation {
					if npayload == nil { // If receiving a end of continuation without expecting one
						c.CloseWithReason(ConnectionCloseReasonProtocolError)
						c.Terminate()
						return 0, nil, ErrProtocolError
					}
					npayload = append(npayload, payload...)
					if nopcode == MessageTypeText && !utf8.Valid(npayload) {
						c.CloseWithReason(ConnectionCloseReasonInconsistentType)
						c.Terminate()
						return 0, nil, encoding.ErrInvalidUTF8
					}
					return nopcode, npayload, nil
				}
				if npayload != nil { // If receiving a non continuation frame expecting one
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, ErrProtocolError
				}
				if opcode == MessageTypeText && !utf8.Valid(payload) {
					c.CloseWithReason(ConnectionCloseReasonInconsistentType)
					c.Terminate()
					return 0, nil, encoding.ErrInvalidUTF8
				}
				return opcode, payload, nil
			}
			if npayload == nil {
				if opcode == MessageTypeContinuation { // If receiving a continuation without without expecting one
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, ErrProtocolError
				}
				npayload = make([]byte, len(payload))
				nopcode = opcode
				copy(npayload[:len(payload)], payload)
			} else {
				if opcode != MessageTypeContinuation { // If receiving a non continuation after sending a prior fragment
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, ErrProtocolError
				}
				lp := len(npayload)
				npayload = append(npayload, make([]byte, len(payload))...)
				copy(npayload[lp:], payload)
			}
		default:
			// Unknown opcode
			c.CloseWithReason(ConnectionCloseReasonProtocolError)
			c.Terminate()
			return 0, nil, ErrProtocolError
		}
	}
}

// ReadMessageTimeout implements the websocket.Connection.ReadMessageTimeout method
func (c *SimpleConnection) ReadMessageTimeout(timeout time.Duration) (MessageType, []byte, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, nil, err
	}
	return c.ReadMessage()
}

// WriteMessage implements the websocket.Connection.WriteMessage method
func (c *SimpleConnection) WriteMessage(opcode MessageType, payload []byte) error {
	return c.WritePacket(byte(opcode), payload)
}

// WriteMessageTimeout implements the websocket.Connection.WriteMessageTimeout method
func (c *SimpleConnection) WriteMessageTimeout(timeout time.Duration, opcode MessageType, payload []byte) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	return c.WriteMessage(opcode, payload)
}
