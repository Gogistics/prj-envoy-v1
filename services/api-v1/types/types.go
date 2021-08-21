package types

import (
    "errors"
)

var (
  ErrNotFound = errors.New("Not found")
)

type APIType struct {
  version string
  lang string
}

type Profile struct {
  Hostname          string
  ServiceProxyIP    string
  Name              string
  Hobbies           []string
}
