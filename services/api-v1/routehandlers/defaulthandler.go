package routehandlers

import (
	"fmt"
	// "log"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/dbhandlers"
	"github.com/Gogistics/prj-envoy-v1/services/api-v1/types"
)

type DefaultWrapper struct{}

var (
	Default DefaultWrapper
)

func countVisitorInfo(req *http.Request) {
	// extract data
	userAgent := req.Header.Get("User-Agent")

	// redis operations
	redisWrapper := dbhandlers.RedisWrapper
	_, errIncrURL := redisWrapper.Incr(req.URL.Path)
	if errIncrURL == nil {
		redisWrapper.Expire(req.URL.Path, 5*time.Minute)
	}
	_, errIncrUserAgent := redisWrapper.Incr(userAgent)
	if errIncrUserAgent == nil {
		redisWrapper.Expire(userAgent, 5*time.Minute)
	}

}

func (wrapper DefaultWrapper) Hello(respWriter http.ResponseWriter, req *http.Request) {
	countVisitorInfo(req)
	remoteAddr, _, err := net.SplitHostPort(req.RemoteAddr)

	// set response
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	profile := types.Profile{
		Hostname:      hostname,
		RemoteAddress: remoteAddr,
		Author:        "Alan Tai",
		Hobbies:       []string{"workout", "programming", "driving"}}

	jProfile, err := json.Marshal(profile)

	if err != nil {
		// handle err
		http.Error(respWriter, err.Error(), http.StatusInternalServerError)
		return
	} else {
		respWriter.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")
		respWriter.Write(jProfile)
	}
}

func (wrapper DefaultWrapper) PostVisitor(respWriter http.ResponseWriter, req *http.Request) {
	countVisitorInfo(req)

	userAgent := req.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "unknown agent"
	}

	// insert visitor infor. into mongo
	req.ParseForm()
	userName := req.Form.Get("userName")
	if userName == "" {
		fmt.Println("unknown user")
	} else {
		fmt.Println("user name: ", userName)

		newData := make(map[string]string)
		newData["user-agent"] = userAgent
		dbhandlers.MongoWrapper.FindOneAndUpdate("userName", userName, &newData)
	}

	// get values through r.Form
	respWriter.WriteHeader(http.StatusCreated)
	respWriter.Write([]byte("Request has been handled successfully"))
}

func (wrapper DefaultWrapper) GetVisitor(respWriter http.ResponseWriter, req *http.Request) {
	countVisitorInfo(req)
	// default return all visitors info from mongo
	mongoWrapper := dbhandlers.MongoWrapper
	// pass key into Find() for sorting order
	// key := "name"
	results := mongoWrapper.Find(nil, nil)

	// get values through r.Form
	respWriter.WriteHeader(http.StatusOK)
	respWriter.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")

	allUsers, err := json.Marshal(results)
	if err != nil {
		// handle err
		http.Error(respWriter, err.Error(), http.StatusInternalServerError)
		return
	} else {
		respWriter.Header().Set("Content-Type", "applicaiton/json; charset=utf-8")
		respWriter.Write(allUsers)
	}
}
