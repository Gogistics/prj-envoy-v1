package main

import (
	"log"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/utilhandlers"
)

func main() {
	var crtPath string = utilhandlers.AppServerHandler.GetCrtPath()
	var keyPath string = utilhandlers.AppServerHandler.GetKeyPath()
	err := utilhandlers.AppServerHandler.GetAppServer().ListenAndServeTLS(crtPath, keyPath)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}
