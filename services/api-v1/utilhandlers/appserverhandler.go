package utilhandlers

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/routehandlers"
	"github.com/gorilla/mux"
)

type appServerHandler struct {
	appRouter *mux.Router
	appMode   bool
	crtPath   string
	keyPath   string
}

var (
	// AppServerHandler is the object handling server config. and init
	AppServerHandler = initAppServerHandler()
	dev              *bool
)

/* Notes
The new router function creates the router and
	returns it to us. We can now use this function
	to instantiate and test the router outside of the main function

ref:
- https://pkg.go.dev/net/http#pkg-constants
- https://pkg.go.dev/flag
*/

func setFlagsVals() {
	dev = flag.Bool("dev", false, "set app mode")
	flag.Parse()
}

func getFlagVal(fg string) interface{} {
	switch fg {
	case "dev":
		return dev
	default:
		log.Println("Warning: flag does not exist!")
		return nil
	}
}

func initAppServerHandler() appServerHandler {
	setFlagsVals()
	appModeInterface := getFlagVal("dev")
	appMode := (*appModeInterface.(*bool))
	var crtPath string
	var keyPath string
	if appMode {
		crtPath = "certs/dev.atai-envoy.com.crt"
		keyPath = "certs/atai-envoy.com.key"
	} else {
		crtPath = "atai-envoy.com.crt"
		keyPath = "atai-envoy.com.key"
	}
	return appServerHandler{appRouter: getNewRouter(), appMode: appMode, crtPath: crtPath, keyPath: keyPath}
}

func getNewRouter() *mux.Router {
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

func (appSH *appServerHandler) GetCrtPath() string {
	return appSH.crtPath
}

func (appSH *appServerHandler) GetKeyPath() string {
	return appSH.keyPath
}

func (appSH *appServerHandler) InitAppServer() *http.Server {
	/* Notes
	Follow Gorilla README to set timeouts to avoid Slowloris attacks.

	ref:
	- https://github.com/gorilla/mux/blob/d07530f46e1eec4e40346e24af34dcc6750ad39f/README.md
	*/
	tlsCfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	appServer := &http.Server{
		Addr:           ":443",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsCfg,
		TLSNextProto:   make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		Handler:        AppServerHandler.appRouter,
	}
	return appServer
}
