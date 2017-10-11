package websocket

import (
	"log"
	"net"
	"time"
)

// SimpleConnection represents a connection with a client
type SimpleConnection struct {
	BaseConnection
}

// NewSimpleConn initialized and return a new websocket.BaseConnection instance
func NewSimpleConn(conn net.Conn) *SimpleConnection {
	return &SimpleConnection{
		BaseConnection: *NewConn(conn),
	}
}

// ReadMessage implements the websocket.Connection.ReadMessage method
func (c *SimpleConnection) ReadMessage() (MessageType, []byte, error) {
	opc, payload, err := c.ReadPacket()
	opcode := MessageType(opc)
	if err != nil {
		return 0, nil, err
	} else if opcode == MessageTypePing {
		if c.state == ConnectionStateOpen {
			log.Println("PING Requested!")
			// Respond the ping message with payload
			c.WritePacketTimeout(time.Millisecond*10, OPCodePongFrame, payload)
		}
		return 0, nil, nil
	} else if opcode == MessageTypePong {
		// TODO Register pong message and save the time of the connection responded
	} else if opcode == MessageTypeConnectionClose {
		c.state = ConnectionStateClosing
		c.Terminate()
		return 0, nil, nil
	}
	return opcode, payload, nil
}

// ReadMessageTimeout implements the websocket.Connection.ReadMessageTimeout method
func (c *SimpleConnection) ReadMessageTimeout(timeout time.Duration) (MessageType, []byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	return c.ReadMessage()
}

// WriteMessage implements the websocket.Connection.WriteMessage method
func (c *SimpleConnection) WriteMessage(opcode MessageType, payload []byte) error {
	return c.WritePacket(byte(opcode), payload)
}

// WriteMessageTimeout implements the websocket.Connection.WriteMessageTimeout method
func (c *SimpleConnection) WriteMessageTimeout(timeout time.Duration, opcode MessageType, payload []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(timeout))
	return c.WriteMessage(opcode, payload)
}
