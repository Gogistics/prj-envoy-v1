package routehandlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebsocketWrapper struct{}

var (
	WebSocket WebsocketWrapper
)

func (wrapper WebsocketWrapper) Communicate(respWriter http.ResponseWriter, req *http.Request) {
	/* Notes
	Ref:
	- https://datatracker.ietf.org/doc/html/rfc6455
	- https://yalantis.com/blog/how-to-build-websockets-in-go/
	- https://github.com/gobwas/ws
	*/
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, errConn := upgrader.Upgrade(respWriter, req, nil)
	if errConn != nil {
		log.Fatalf("Error: WS failed to build connection. ", errConn)
		return
	}

	for {
		msgType, msg, errReadMsg := conn.ReadMessage()
		if errReadMsg != nil {
			return
		}

		// print msg
		log.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

		// Write msg back to client
		if errWriteMsg := conn.WriteMessage(msgType, msg); errWriteMsg != nil {
			return
		}
	}
}
