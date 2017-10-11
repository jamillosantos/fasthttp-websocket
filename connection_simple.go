package websocket

import (
	"net"
	"time"
	"log"
)

// BaseConnection represents a connection with a client
type SimpleConnection struct {
	BaseConnection
}

// NewConn initialized and return a new websocket.BaseConnection instance
func NewSimpleConn(conn net.Conn) *SimpleConnection {
	return &SimpleConnection{
		BaseConnection: *NewConn(conn),
	}
}

func (c *BaseConnection) ReadMessage() (MessageType, []byte, error) {
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

func (c *BaseConnection) ReadMessageTimeout(timeout time.Duration) (MessageType, []byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	return c.ReadMessage()
}

func (c *BaseConnection) WriteMessage(opcode MessageType, payload []byte) error {
	return c.WritePacket(byte(opcode), payload)
}

func (c *BaseConnection) WriteMessageTimeout(timeout time.Duration, opcode MessageType, payload []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(timeout))
	return c.WriteMessage(opcode, payload)
}
