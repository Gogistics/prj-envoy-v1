package routehandlers

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/dbhandlers"
	"github.com/Gogistics/prj-envoy-v1/services/api-v1/grpchandlers"
	"github.com/Gogistics/prj-envoy-v1/services/api-v1/types"
)

type DefaultHandler struct{}

var (
	// Default object of handling default routes
	Default      DefaultHandler
	redisWrapper = dbhandlers.RedisWrapper
)

func countVisitorInfo(req *http.Request) {
	// extract data
	userAgent := req.Header.Get("User-Agent")

	// redis operations
	_, errIncrURL := redisWrapper.Incr(req.URL.Path)
	if errIncrURL == nil {
		redisWrapper.Expire(req.URL.Path, 5*time.Minute)
	}
	_, errIncrUserAgent := redisWrapper.Incr(userAgent)
	if errIncrUserAgent == nil {
		redisWrapper.Expire(userAgent, 5*time.Minute)
	}

}

func (handler DefaultHandler) HelloGRPC(respWriter http.ResponseWriter, req *http.Request) {
	countVisitorInfo(req)

	// TODO: handle errors and add timeout
	grpchandlers.GRPCWrapper.RunGRPCCalls()

	respWriter.WriteHeader(http.StatusAccepted)
	respWriter.Write([]byte("Request has been accepted and in processing"))
}

func (handler DefaultHandler) Hello(respWriter http.ResponseWriter, req *http.Request) {
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
		respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
		respWriter.Write(jProfile)
	}
}

func (handler DefaultHandler) PostVisitor(respWriter http.ResponseWriter, req *http.Request) {
	countVisitorInfo(req)

	userAgent := req.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "unknown agent"
	}

	// insert visitor infor. into mongo
	req.ParseForm()
	userName := req.Form.Get("userName")
	if userName == "" {
		log.Println("unknown user")
	} else {
		log.Println("user name: ", userName)

		newData := make(map[string]string)
		newData["user-agent"] = userAgent
		dbhandlers.MongoWrapper.FindOneAndUpdate("userName", userName, &newData)
	}

	respWriter.WriteHeader(http.StatusCreated)
	respWriter.Write([]byte("Request has been handled successfully"))
}

func (handler DefaultHandler) GetVisitor(respWriter http.ResponseWriter, req *http.Request) {
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
