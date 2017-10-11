package websocket

import (
	"github.com/pkg/errors"
	"sync"
	"time"
)

// ListenableManager is a websocket.Manager that implements a set of handlers
// that will be called when any events occurs
type ListenableManager struct {
	ReadTimeout    time.Duration
	conns          sync.Pool
	OnConnect      ConnectionHandler
	OnMessage      MessageHandler
	OnMessageError ConnectionErrorHandler
	OnClose        ConnectionHandler
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
		err = cm.OnConnect(c)
		if err != nil {
			err2 := c.Conn().Close()
			if err2 != nil {
				return errors.Wrap(err2, err.Error())
			}
			return err
		}
	}
	defer func() {
		if r := recover(); r != nil {
			err2 := c.CloseWithReason(ConnectionCloseReasonUnexpected)
			switch x := r.(type) {
			case string:
				if err2 != nil {
					err = errors.Wrap(err2, x)
				} else {
					err = errors.New(x)
				}
			case error:
				if err2 != nil {
					err = errors.Wrap(err2, x.Error())
				} else {
					err = x
				}
			default:
				err = errors.New("unknown error")
			}
		}
	}()
	for !c.IsClosed() {
		opcode, payload, err := c.ReadMessageTimeout(cm.ReadTimeout)
		if err == nil && payload != nil {
			err = cm.OnMessage(c, opcode, payload)
			if err != nil && cm.OnMessageError != nil {
				cm.OnMessageError(c, err)
			}
		}
	}
	return cm.OnClose(c)
}
