package websocket

import "sync"

// SimpleManager is a manager that will let the handler property manage all
// reading and writing of the connection.
type SimpleManager struct {
	conns   sync.Pool
	handler ConnectionHandler
}

// NewSimpleManager creates a new instance of the SimpleManager
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

// Accept implements the websocket.Manager.Accept method
func (cm *SimpleManager) Accept(ctx *ConnectionContext) error {
	c := cm.conns.Get().(*SimpleConnection)
	c.Init(ctx)
	return cm.handler(c)
}
