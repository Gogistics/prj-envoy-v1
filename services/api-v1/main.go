package main

import (
	"log"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/utilhandlers"
)

/* Notes
- entry of the web app
*/
func main() {
	err := utilhandlers.AppServerHandler.InitAppServer()
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}
