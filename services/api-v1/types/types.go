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
  RemoteAddress    string
  Author              string
  Hobbies           []string
}
