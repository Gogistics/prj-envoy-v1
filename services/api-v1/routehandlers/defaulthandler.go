package routehandlers

import (
  "net/http"
)

type DefaultWrapper struct

var (
  Default DefaultWrapper
)

func countUserAgent(req *http.Request) {
  // extract data
  ip, _, err := net.SplitHostPort(r.RemoteAddr)
  forward := r.Header.Get("X-Forwarded-For")
  userAgent := r.Header.Get("User-Agent")
}

func (wrapper DefaultWrapper) Hello(respWriter http.ResponseWriter, req *http.Request) {
  countUserAgent(req)
}

func (wrapper DefaultWrapper) HandleVisitor(respWriter http.ResponseWriter, req *http.Request) {
  countUserAgent(req)
}

func (wrapper DefaultWrapper) GetVisitorInfo(respWriter http.ResponseWriter, req *http.Request) {
  countUserAgent(req)
  // default return all visitors
}