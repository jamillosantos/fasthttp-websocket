package main

import (
	"github.com/valyala/fasthttp"
	"github.com/jamillosantos/websocket"
	"log"
	"time"
)

type BuffWriter struct {
	Buff []byte
}

func main() {
	server := &fasthttp.Server{}
	manager := websocket.NewListeableManager()
	manager.ReadTimeout = time.Second * 10
	manager.OnConnect = func(conn websocket.Connection) error {
		log.Println("Incoming client at", conn.Conn().RemoteAddr())
		return nil
	}
	manager.OnMessage = func(conn websocket.Connection, opcode websocket.MessageType, payload []byte) error {
		log.Println("OnMessage", opcode, payload)
		return nil
	}
	manager.OnClose = func(conn websocket.Connection) error {
		log.Println("see ya", conn.Conn().RemoteAddr())
		return nil
	}
	upgrader := websocket.NewUpgrader(manager)
	/*
	upgrader := websocket.NewUpgrader(websocket.NewSimpleManager(func(conn websocket.Connection) error {
		for {
			opcode, data, err := conn.ReadMessageTimeout(time.Millisecond * 10)
			if err != nil && data == nil {
				log.Println(err)
			} else if data != nil {
				switch opcode {
				case websocket.MessageTypeText:
					log.Println(string(data))
					break
				case websocket.MessageTypeBinary:
					log.Println(data)
					break
				case websocket.MessageTypeContinuation:
					log.Println("Cont", string(data))
					break
				}
			}
			conn.WritePacket(websocket.OPCodeTextFrame, []byte(time.Now().String()))
			time.Sleep(time.Second)
		}
	}))
	*/

	server.Handler = func(ctx *fasthttp.RequestCtx) {
		upgrader.Upgrade(ctx)
	}

	server.ListenAndServe(":8080")
}
