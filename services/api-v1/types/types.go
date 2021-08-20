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
  Name        string
  Hostname    string
  Hobbies     []string
}
