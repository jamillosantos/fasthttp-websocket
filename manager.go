package websocket

import (
	"net"
	"sync"
	"time"
	"errors"
)

// ConnectionHandler represents the handler that the Upgrader will trigger after
// successfully upgrading the connection.
type ConnectionHandler func(conn Connection) error

type MessageHandler func(conn Connection, opcode MessageType, payload []byte) error

type ConnectionContext struct {
	Conn       net.Conn
	Compressed bool
}

type Manager interface {
	Accept(conn *ConnectionContext) error
}

type Message struct {
	opcode  byte
	payload []byte
}

type SimpleManager struct {
	conns   sync.Pool
	handler ConnectionHandler
}

type ListenableManager struct {
	conns       sync.Pool
	ReadTimeout time.Duration
	OnConnect   ConnectionHandler
	OnMessage   MessageHandler
	OnClose     ConnectionHandler
}

func NewSimpleManager(handler ConnectionHandler) *SimpleManager {
	return &SimpleManager{
		conns: sync.Pool{
			New: func() interface{} {
				return NewConn(nil)
			},
		},
		handler: handler,
	}
}

func NewListeableManager() *ListenableManager {
	return &ListenableManager{
		conns: sync.Pool{
			New: func() interface{} {
				return NewConnectionListable(nil)
			},
		},
	}
}

func (cm *SimpleManager) Accept(ctx *ConnectionContext) error {
	c := cm.conns.Get().(*BaseConnection)
	c.Init(ctx)
	return cm.handler(c)
}

func (cm *ListenableManager) Accept(ctx *ConnectionContext) (err error) {
	c := cm.conns.Get().(*ListenableConnection)
	defer func() {
		c.Reset()
		cm.conns.Put(c)
	}()
	c.Init(ctx)
	if cm.OnConnect != nil {
		if err := cm.OnConnect(c); err != nil {
			c.conn.Close()
			return err
		}
	}
	defer func() {
		if r := recover(); r != nil {
			c.CloseWithReason(ConnectionCloseReasonUnexpected)
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown error")
			}
		}
	}()
	for !c.IsClosed() {
		opcode, payload, err := c.ReadMessageTimeout(cm.ReadTimeout)
		if err == nil && payload != nil {
			cm.OnMessage(c, opcode, payload)
		}
	}
	return cm.OnClose(c)
}
