package websocket

import (
	"net"
)

// ConnectionHandler represents the handler that the Upgrader will trigger after
// successfully upgrading the connection.
type ConnectionHandler func(conn Connection) error

// MessageHandler represents the handler for a message.
type MessageHandler func(conn Connection, opcode MessageType, payload []byte) error

// ConnectionContext saves all the data that will be forwarded to the manager
// from the hijacked connection.
type ConnectionContext struct {
	Conn       net.Conn
	Compressed bool
}

// Manager handles all the tasks .
type Manager interface {
	// Accept handles the incoming connection.
	Accept(conn *ConnectionContext) error
}
