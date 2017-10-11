package websocket

import (
	"errors"
	"sync"
	"time"
)

// ListenableManager is a websocket.Manager that implements a set of handlers
// that will be called when any events occurs
type ListenableManager struct {
	conns       sync.Pool
	ReadTimeout time.Duration
	OnConnect   ConnectionHandler
	OnMessage   MessageHandler
	OnClose     ConnectionHandler
}

// NewListeableManager returns a new instance of the websocket.ListenableManager
func NewListeableManager() *ListenableManager {
	return &ListenableManager{
		conns: sync.Pool{
			New: func() interface{} {
				return NewSimpleConn(nil)
			},
		},
	}
}

// Accept implements the websocket.Manager.Accept method
func (cm *ListenableManager) Accept(ctx *ConnectionContext) (err error) {
	c := cm.conns.Get().(Connection)
	defer func() {
		c.Reset()
		cm.conns.Put(c)
	}()
	c.Init(ctx)
	if cm.OnConnect != nil {
		if err := cm.OnConnect(c); err != nil {
			c.Conn().Close()
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
