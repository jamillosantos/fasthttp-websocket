package main

import (
	"github.com/jamillosantos/websocket"
	"github.com/valyala/fasthttp"
	"log"
	"time"
	"fmt"
)

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
	server.Handler = func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.URI().Path()) {
		case "/ws":
			upgrader.Upgrade(ctx)
		case "/":
			fmt.Fprint(ctx, "This is the root of the server")
		default:
			fmt.Fprint(ctx, "404 Not Found")
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}

	server.ListenAndServe(":8080")
}
