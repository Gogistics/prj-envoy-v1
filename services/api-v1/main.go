package main

import (
  "fmt"
  "log"
  "os"
  "encoding/json"
  "net/http"
  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"
  "github.com/Gogistics/prj-envoy-v1/services/api-v1/types"
)

// The new router function creates the router and
// returns it to us. We can now use this function
// to instantiate and test the router outside of the main function
func newRouter() *mux.Router {
  r := mux.NewRouter()
  r.HandleFunc("/api/v1", handlerHello).Methods("GET")
  r.HandleFunc("/ws-echo", handlerWS)
  return r
}

func main() {
  // The router is now formed by calling the `newRouter` constructor function
  // that we defined above. The rest of the code stays the same
  r := newRouter()
  err := http.ListenAndServeTLS(":443", "atai-envoy.com.crt", "atai-envoy.com.key", r)
  if err != nil {
    log.Fatal("ListenAndServeTLS: ", err)
  }
}

func handlerHello(w http.ResponseWriter, r *http.Request) {
  hostname, err := os.Hostname()
  if err != nil {
    panic(err)
  }

  profile := types.Profile{Name: "Alan", Hostname: hostname, Hobbies: []string{"workout", "programming", "driving"}}
  jProfile, err := json.Marshal(profile)

  if err != nil {
    // handle err
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")
  w.Write(jProfile) 
}

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
