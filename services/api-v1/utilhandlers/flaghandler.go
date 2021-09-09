package utilhandlers

import (
	"flag"
	"log"
)

type appFlagHandler struct{}

var (
	FlagHandler = getFlagHandler()
	dev         *bool
	sessionKey  *string
)

func getFlagHandler() appFlagHandler {
	dev = flag.Bool("dev", false, "set app mode")
	sessionKey = flag.String("sessionKey", "", "set Redis seesion store")
	flag.Parse()
	return appFlagHandler{}
}

func (fgHandler *appFlagHandler) GetFlagVal(fg string) interface{} {
	switch fg {
	case "dev":
		return dev
	case "sessionKey":
		return sessionKey
	default:
		log.Println("Warning: flag does not exist!")
		return nil
	}
}
