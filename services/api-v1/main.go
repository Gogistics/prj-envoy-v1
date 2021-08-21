package main

import (
  "fmt"
  "log"
  "os"
  "net"
  "encoding/json"
  "net/http"
  "flag"
  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"
  "github.com/Gogistics/prj-envoy-v1/services/api-v1/types"
  "github.com/Gogistics/prj-envoy-v1/services/api-v1/dbhandlers"
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

func handlerHello(w http.ResponseWriter, r *http.Request) {
  ip, _, err := net.SplitHostPort(r.RemoteAddr)
  forward := r.Header.Get("X-Forwarded-For")
  
  // redis wrapper
  redisWrapper := dbhandlers.RedisWrapper
  redisWrapper.Set("ip", []byte(ip))
  redisWrapper.Set("forward", []byte(forward))

  // set response
  hostname, err := os.Hostname()
  if err != nil {
    panic(err)
  }

  ipFromRedis, errRedisGet := redisWrapper.Get("ip")
  if errRedisGet != nil {
    fmt.Errorf("error on Redis Get: %s", errRedisGet)
    ipFromRedis = []byte("NA")
  }

  profile := types.Profile{
    Hostname: hostname,
    ServiceProxyIP: string(ipFromRedis),
    Name: "Alan",
    Hobbies: []string{"workout", "programming", "driving"}}

  jProfile, err := json.Marshal(profile)

  if err != nil {
    // handle err
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")
  w.Write(jProfile) 
}

// TODO: complete WS
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
