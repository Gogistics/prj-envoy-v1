package main

import (
  "fmt"
  "log"
  "os"
  "net"
  "encoding/json"
  "net/http"
  "flag"
  "time"
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
  // https://pkg.go.dev/net/http#pkg-constants
  r.HandleFunc("/api/v1", handlerHello).Methods(http.MethodGet)
  r.HandleFunc("/api/v1/visitor", handlerVisitorPost).Methods(http.MethodPost)
  r.HandleFunc("/api/v1/visitor", handlerVisitorGet).Methods(http.MethodGet)
  r.HandleFunc("/ws-echo", handlerWS)
  return r
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


// TODO: move to routehandlers
func handlerHello(w http.ResponseWriter, r *http.Request) {
  // extract data from http header
  remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
  userAgent := r.Header.Get("User-Agent")

  // redis operations
  redisWrapper := dbhandlers.RedisWrapper
  _, errIncrURL := redisWrapper.Incr(r.URL.Path)
  if errIncrURL == nil {
    redisWrapper.Expire(r.URL.Path, 5 * time.Minute)
  }
  _, errIncrUserAgent := redisWrapper.Incr(userAgent)
  if errIncrUserAgent == nil {
    redisWrapper.Expire(userAgent, 5 * time.Minute)
  }

  // test []byte
  redisWrapper.Set("remoteAddr", []byte(remoteAddr))

  // set response
  hostname, err := os.Hostname()
  if err != nil {
    panic(err)
  }

  profile := types.Profile{
    Hostname: hostname,
    RemoteAddress: remoteAddr,
    Author: "Alan Tai",
    Hobbies: []string{"workout", "programming", "driving"}}

  jProfile, err := json.Marshal(profile)

  if err != nil {
    // handle err
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  } else {
    w.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")
    w.Write(jProfile) 
  }
}

// curl -d "userName=alan" -k -X POST https://0.0.0.0/api/v1/visitor
func handlerVisitorPost(w http.ResponseWriter, r *http.Request) {
  redisWrapper := dbhandlers.RedisWrapper
  _, errIncr := redisWrapper.Incr(r.URL.Path)
  if errIncr == nil {
    redisWrapper.Expire(r.URL.Path, 5 * time.Minute)
  }

  userAgent := r.Header.Get("User-Agent")
  if userAgent == "" {
    userAgent = "unknown agent"
  }

  // insert visitor infor. into mongo
  r.ParseForm()
  userName := r.Form.Get("userName")
  if userName == "" {
    fmt.Println("unknown user")
  } else {
    fmt.Println("user name: ", userName)

    newData := make(map[string]string)
    newData["user-agent"] = userAgent
    dbhandlers.MongoWrapper.FindOneAndUpdate("userName", userName, &newData)
  }

  // get values through r.Form
  w.WriteHeader(http.StatusCreated)
  w.Write([]byte("Request has been handled successfully"))
}

func handlerVisitorGet(w http.ResponseWriter, r *http.Request) {
  redisWrapper := dbhandlers.RedisWrapper
  _, errIncr := redisWrapper.Incr(r.URL.Path)
  if errIncr == nil {
    redisWrapper.Expire(r.URL.Path, 5 * time.Minute)
  }
  // get all visitors infor. from mongo
  mongoWrapper := dbhandlers.MongoWrapper
  // pass key into Find() for sorting order
  // key := "name"
  results := mongoWrapper.Find(nil, nil)

  // get values through r.Form
  w.WriteHeader(http.StatusOK)
  w.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")

  allUsers, err := json.Marshal(results)
  if err != nil {
    // handle err
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  } else {
    w.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")
    w.Write(allUsers) 
  }
}
// \move to routehandlers


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
