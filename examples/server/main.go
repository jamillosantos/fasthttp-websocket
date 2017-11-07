package main

import (
	"github.com/jamillosantos/websocket"
	"github.com/valyala/fasthttp"
	"log"
	"time"
)

func main() {
	server := &fasthttp.Server{}
	manager := websocket.NewListeableManager()
	manager.ReadTimeout = time.Second * 10
	manager.OnMessageError = func(conn websocket.Connection, err error) {
		log.Println(err)
	}
	manager.OnConnect = func(conn websocket.Connection) error {
		log.Println("Incoming client at", conn.Conn().RemoteAddr())
		return nil
	}
	manager.OnMessage = func(conn websocket.Connection, opcode websocket.MessageType, payload []byte) error {
		log.Println("OnMessage", opcode, payload)
		conn.WriteMessage(opcode, payload)
		return nil
	}
	manager.OnClose = func(conn websocket.Connection) error {
		log.Println("see ya", conn.Conn().RemoteAddr())
		return nil
	}
	upgrader := websocket.NewUpgrader(manager)
	upgrader.Error = func(ctx *fasthttp.RequestCtx, reason error) {
		log.Println(reason)
	}
	server.Handler = func(ctx *fasthttp.RequestCtx) {
		log.Println("Connection received.")
		upgrader.Upgrade(ctx)
		/*
		switch string(ctx.URI().Path()) {
		case "/ws":
			upgrader.Upgrade(ctx)
		case "/":
			fmt.Fprint(ctx, "This is the root of the server")
		default:
			fmt.Fprint(ctx, "404 Not Found")
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
		*/
	}

	server.ListenAndServe(":9001")
}
