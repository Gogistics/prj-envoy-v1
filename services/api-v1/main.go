package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/routehandlers"
	"github.com/gorilla/mux"
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
	rtr.HandleFunc("/ws/v1", routehandlers.WebSocket.Communicate)
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
