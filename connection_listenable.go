package websocket

import "net"

type ConnectionListener func(c Connection, opcode MessageType, payload []byte)

type ListenableConnection struct {
	BaseConnection
	listener ConnectionListener
}

func NewConnectionListable(conn net.Conn) *ListenableConnection {
	return &ListenableConnection{
		BaseConnection: *NewConn(conn),
	}
}
