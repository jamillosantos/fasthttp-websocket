package websocket

import (
	"net"
	"time"
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
		case MessageTypePing:
			if c.state == ConnectionStateOpen {
				if len(payload) > 125 {
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, nil
				}
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
			c.state = ConnectionStateClosing
			c.Close()
			err = c.Terminate()
			return 0, nil, err
		case MessageTypeContinuation, MessageTypeBinary, MessageTypeText:
			if fin {
				if opcode == MessageTypeContinuation {
					if npayload == nil {
						c.CloseWithReason(ConnectionCloseReasonProtocolError)
						c.Terminate()
						return 0, nil, ErrProtocolError
					}
					return nopcode, append(npayload, payload...), nil
				}
				if npayload != nil {
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, ErrProtocolError
				}
				return opcode, payload, nil
			}
			if npayload == nil {
				if opcode == MessageTypeContinuation {
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, ErrProtocolError
				}
				npayload = make([]byte, len(payload))
				nopcode = opcode
				copy(npayload[:len(payload)], payload)
			} else {
				if opcode != MessageTypeContinuation {
					c.CloseWithReason(ConnectionCloseReasonProtocolError)
					c.Terminate()
					return 0, nil, ErrProtocolError
				}
				lp := len(npayload)
				npayload = append(npayload, make([]byte, len(payload))...)
				copy(npayload[lp:], payload)
			}
		default:
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
