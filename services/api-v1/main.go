package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/routehandlers"
	"github.com/gorilla/mux"
)

/* Notes:
ref:
- https://pkg.go.dev/net/http#pkg-constants
- https://pkg.go.dev/flag
*/

// Note: The new router function creates the router and
// 	returns it to us. We can now use this function
// 	to instantiate and test the router outside of the main function
func newRouter() *mux.Router {
	rtr := mux.NewRouter()
	// general REST APIs
	rtr.HandleFunc("/api/v1", routehandlers.Default.Hello).Methods(http.MethodGet)
	rtr.HandleFunc("/api/v1/visitor", routehandlers.Default.PostVisitor).Methods(http.MethodPost)
	rtr.HandleFunc("/api/v1/visitor", routehandlers.Default.GetVisitor).Methods(http.MethodGet)

	// TODO: complete websocket
	rtr.HandleFunc("/ws/v1", routehandlers.WebSocket.Communicate)

	// gRPC
	rtr.HandleFunc("/grpc/v1", routehandlers.Default.HelloGRPC).Methods(http.MethodPost)
	rtr.NotFoundHandler = rtr.NewRoute().HandlerFunc(http.NotFound).GetHandler()
	return rtr
}

func main() {
	var dev = flag.Bool("dev", false, "set app mode")
	flag.Parse()

	// The router is now formed by calling the `newRouter` constructor function
	// that we defined above. The rest of the code stays the same
	appRouter := newRouter()

	// TODO: move tls config. to config. handler
	var crtPath string
	var keyPath string
	if *dev {
		crtPath = "certs/dev.atai-envoy.com.crt"
		keyPath = "certs/atai-envoy.com.key"
	} else {
		crtPath = "atai-envoy.com.crt"
		keyPath = "atai-envoy.com.key"
	}

	/* Notes
	Follow Gorilla README to set timeouts to avoid Slowloris attacks.
	ref:
	- https://github.com/gorilla/mux/blob/d07530f46e1eec4e40346e24af34dcc6750ad39f/README.md
	*/
	appServer := &http.Server{
		Addr:           ":443",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        appRouter,
	}

	err := appServer.ListenAndServeTLS(crtPath, keyPath)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}
