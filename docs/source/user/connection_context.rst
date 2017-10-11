Connection Session
==================

Sometimes we need to attach some information to a connection that just started.
In order to provide this functionality the ``Connection interface`` provides a
``Context()`` and ``SetContext()`` methods.

Example
-------
.. code-block:: go
	:linenos:
	:emphasize-lines: 10-12, 19-21, 25-26

	package main

	import (
		"github.com/jamillosantos/websocket"
		"github.com/valyala/fasthttp"
		"fmt"
		"log"
	)

	type ConnCtx struct {
		name string
	}

	func main() {
		server := &fasthttp.Server{}
		manager := websocket.NewListeableManager()
		manager.OnConnect = func(conn websocket.Connection) error {
			log.Println("Incoming client ", conn.Conn().RemoteAddr())
			conn.SetContext(&ConnCtx{
				name: "John Doe",
			})
			return nil
		}
		manager.OnMessage = func(conn websocket.Connection, opcode websocket.MessageType, payload []byte) error {
			ctx := conn.Context().(*ConnCtx)
			log.Println("message from", ctx.name, opcode, payload)
			return nil
		}
		manager.OnClose = func(conn websocket.Connection) error {
			log.Println("see ya", conn.Conn().RemoteAddr())
			return nil
		}
		upgrader := websocket.NewUpgrader(manager)
		server.Handler = func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.URI().Path()) {
			case "/":
				fmt.Fprint(ctx, "This is the root of the server")
			case "/ws":
				upgrader.Upgrade(ctx)
			default:
				fmt.Fprint(ctx, "404 Not Found")
				ctx.SetStatusCode(fasthttp.StatusNotFound)
			}
		}

		server.ListenAndServe(":8080")
	}
..


