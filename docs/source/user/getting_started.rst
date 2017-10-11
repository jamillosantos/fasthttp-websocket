Getting started
===============

This section will take you through the installation and your first
server implementation.

Installation
------------

To start using the library you must fetch it from the github using th ``go get``
command::

    $ go get github.com/jamillosantos/websocket

Usage
-----

This section will show how the WebSocket library will be integrated to your
existing fasthttp server.

Starting with a basic integration, the modifications on the code will be added
in steps until the server is done.

1. A simple fasthttp server
+++++++++++++++++++++++++++

This is a simple code that will serve a message when the client access the root
endpoint (/).

.. code-block:: go
   :linenos:

   package main

   import (
       "github.com/valyala/fasthttp"
       "fmt"
   )

   func main() {
       server := &fasthttp.Server{}
       server.Handler = func(ctx *fasthttp.RequestCtx) {
           switch string(ctx.URI().Path()) {
           case "/":
               fmt.Fprint(ctx, "This is the root of the server")
           default:
               fmt.Fprint(ctx, "404 Not Found")
               ctx.SetStatusCode(fasthttp.StatusNotFound)
           }
       }

       server.ListenAndServe(":8080")
   }
::

2. Setting the WebSocket library
++++++++++++++++++++++++++++++++

With a few line line added, you will get your application responding to websocket
connections.

.. code-block:: go
   :linenos:
   :emphasize-lines: 4,11-24, 29-30

   package main

   import (
       "github.com/jamillosantos/websocket"
       "github.com/valyala/fasthttp"
       "fmt"
   )

   func main() {
       server := &fasthttp.Server{}
       manager := websocket.NewListeableManager()
       manager.OnConnect = func(conn websocket.Connection) error {
           log.Println("Incoming client ", conn.Conn().RemoteAddr())
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
