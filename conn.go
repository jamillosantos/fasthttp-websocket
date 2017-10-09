package websocket

import (
	"net"
	"time"
	"io"
)

// Conn represents a connection with a client
type Conn struct {
	readBuff   []byte
	conn       net.Conn
	compressed bool
}

// NewConn initialized and return a new websocket.Conn instance
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		readBuff: make([]byte, 1024*8),
		conn:     conn,
	}
}

// Reset cleans up all the data and prepare the instance for being placed back
// on the pool, for avoiding allocation.
func (c *Conn) Reset() {
}

func (c *Conn) ReadMessage(buff []byte, deadline time.Time) (int, error) {
	// TODO to implement this method.
	c.conn.SetReadDeadline(deadline)
	n, err := c.conn.Read(c.readBuff)
	if err == io.EOF {
		err = nil
	}
	return n, nil
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *Conn) preparePacket(opcode byte, payload []byte) ([]byte, error) {
	return EncodePacket(true, c.compressed, false, false, opcode, uint64(len(payload)), nil, payload)
}

func (c *Conn) WriteTextMessage(data []byte) error {
	packet, err := c.preparePacket(OPCodeTextFrame, data)
	if err != nil {
		return err
	}
	_, err = c.Write(packet)
	return err
}
