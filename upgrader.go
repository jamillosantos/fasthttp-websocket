package websocket

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
	"net"
)

var (
	globalUID                 = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	strUpgrade                = []byte("Upgrade")
	strConnection             = []byte("Connection")
	strwebsocket              = []byte("websocket")
	strSecWebSocketAccept     = []byte("Sec-WebSocket-Accept")
	strSecWebSocketKey        = []byte("Sec-WebSocket-Key")
	strSecWebSocketVersion    = []byte("Sec-WebSocket-Version")
	strSecWebSocketVersion13  = []byte("13")
	strSecWebSocketProtocol   = []byte("Sec-WebSocket-Protocol")
	strSecWebSocketExtensions = []byte("Sec-WebSocket-Extensions")
	strPerMessageDeflate      = []byte("permessage-deflate")
)

// HandshakeError represents an handshake error while upgrading a connection.
type HandshakeError struct {
	message string
}

func (e HandshakeError) Error() string {
	return e.message
}

// Upgrader implements build the HTTP Package for upgrading the connection from
// regular HTTP Request to a Websocket request.
type Upgrader struct {
	manager Manager
	Error   func(ctx *fasthttp.RequestCtx, reason error)
}

// NewUpgrader returns a new instance of an websocket.Upgrader
func NewUpgrader(manager Manager) *Upgrader {
	return &Upgrader{
		manager: manager,
	}
}

func (u *Upgrader) reportError(ctx *fasthttp.RequestCtx, status int, reason string) error {
	err := HandshakeError{reason}
	ctx.Response.SetStatusCode(status)
	if u.Error != nil {
		u.Error(ctx, err)
	} else {
		ctx.Response.Header.Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(ctx, reason)
	}
	return err
}

// Upgrade upgrades the request to the websocket protocol
//
// TODO To document
func (u *Upgrader) Upgrade(ctx *fasthttp.RequestCtx) error {
	if !ctx.IsGet() {
		return u.reportError(ctx, fasthttp.StatusMethodNotAllowed, "Method not allowed")
	}

	if !bytes.Equal(ctx.Request.Header.PeekBytes(strConnection), strUpgrade) {
		return u.reportError(ctx, fasthttp.StatusBadRequest, "Invalid connection type")
	}

	upgradeTo := ctx.Request.Header.PeekBytes(strUpgrade)
	if !bytes.Equal(upgradeTo, strwebsocket) {
		return u.reportError(ctx, fasthttp.StatusBadRequest, fmt.Sprintf("This connection cannot be upgraded to '%s'", upgradeTo))
	}

	key := ctx.Request.Header.PeekBytes(strSecWebSocketKey)
	if key == nil {
		return u.reportError(ctx, fasthttp.StatusBadRequest, "The key is missing.")
	}

	version := ctx.Request.Header.PeekBytes(strSecWebSocketVersion)
	if version == nil {
		return u.reportError(ctx, fasthttp.StatusBadRequest, "No version provided.")
	}
	if !bytes.Equal(version, strSecWebSocketVersion13) {
		return u.reportError(ctx, fasthttp.StatusBadRequest, "The version is not supported.")
	}

	compress := false
	headerVisit(ctx.Request.Header.PeekBytes(strSecWebSocketExtensions), func(k, v []byte) bool {
		log.Println(string(v))
		if bytes.Equal(v, strPerMessageDeflate) {
			compress = true
			return false
		}
		return true
	})

	// TODO: Check origin

	ctx.Response.SetStatusCode(fasthttp.StatusSwitchingProtocols)
	ctx.Response.Header.AddBytesKV(strUpgrade, strwebsocket)
	ctx.Response.Header.AddBytesKV(strConnection, strUpgrade)
	ctx.Response.Header.AddBytesKV(strSecWebSocketAccept, generateAcceptFromKey(key))

	if compress {
		ctx.Response.Header.AddBytesK(strSecWebSocketExtensions, "permessage-deflate; server_no_context_takeover; client_no_context_takeover")
	} else {
		ctx.Response.Header.AddBytesK(strSecWebSocketExtensions, "server_no_context_takeover; client_no_context_takeover")
	}

	ctx.Hijack(func(c net.Conn) {
		u.manager.Accept(&ConnectionContext{
			Compressed: compress,
			Conn:       c,
		})
	})
	return nil
}

func generateAcceptFromKey(key []byte) []byte {
	s := sha1.New()
	s.Write(key)
	s.Write(globalUID)
	data := s.Sum(nil)
	result := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(result, data)
	return result
}

func headerVisit(header []byte, f func(name, value []byte) bool) {
	l := len(header)
	var (
		found bool
		bAt   int
	)
	for i := 0; i < l; i++ {
		b := header[i]
		if (b != ' ') && (b != '\t') {
			bAt = i
			found = false
			for i < l {
				b = header[i]
				if b == ';' {
					if !f(header[bAt:bAt], header[bAt:i]) {
						return
					}
					found = true
					break
				} else if b == '=' {
					name := header[bAt:i]
					i++
					for i < l {
						b = header[i]
						if b != ' ' && b != '\t' {
							break
						}
						i++
					}
					bAt = i
					for i < l {
						if header[i] == ';' {
							break
						}
						i++
					}
					if !f(name, header[bAt:i]) {
						return
					}
					found = true
					break
				}
				i++
			}
			if !(found || f([]byte(""), header[bAt:i])) {
				return
			}
		}
	}
}
