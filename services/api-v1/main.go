package main

import (
  "fmt"
  "log"
  "net/http"
  "flag"
  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"
  "github.com/Gogistics/prj-envoy-v1/services/api-v1/routehandlers"
)

// The new router function creates the router and
// returns it to us. We can now use this function
// to instantiate and test the router outside of the main function
func newRouter() *mux.Router {
  rtr := mux.NewRouter()
  // https://pkg.go.dev/net/http#pkg-constants
  rtr.HandleFunc("/api/v1", routehandlers.Default.Hello).Methods(http.MethodGet)
  rtr.HandleFunc("/api/v1/visitor", routehandlers.Default.PostVisitor).Methods(http.MethodPost)
  rtr.HandleFunc("/api/v1/visitor", routehandlers.Default.GetVisitor).Methods(http.MethodGet)
  rtr.HandleFunc("/ws-echo", handlerWS)
  rtr.NotFoundHandler = rtr.NewRoute().HandlerFunc(http.NotFound).GetHandler()
  return rtr
}

func main() {
  // https://pkg.go.dev/flag
  dev := flag.Bool("dev", false, "set app mode")
  flag.Parse()

  // The router is now formed by calling the `newRouter` constructor function
  // that we defined above. The rest of the code stays the same
  r := newRouter()
  var crtPath string
  var keyPath string
  if *dev {
    crtPath = "certs/dev.atai-envoy.com.crt"
    keyPath = "certs/atai-envoy.com.key"
  } else {
    crtPath = "atai-envoy.com.crt"
    keyPath = "atai-envoy.com.key"
  }
  err := http.ListenAndServeTLS(":443", crtPath, keyPath, r)
  if err != nil {
    log.Fatal("ListenAndServeTLS: ", err)
  }
}


// TODO: move to routehandlers and complete WS
func handlerWS(w http.ResponseWriter, r *http.Request) {
  var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
  }

  conn, errConn := upgrader.Upgrade(w, r, nil)
  if errConn != nil {
    log.Fatal("WS failed to build connection")
    return
  }

  for {
    msgType, msg, errReadMsg := conn.ReadMessage()
    if errReadMsg != nil {
      return
    }

    // print msg
    fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

    // Write msg back to client
    if errWriteMsg := conn.WriteMessage(msgType, msg); errWriteMsg != nil {
        return
    }
  }
}
