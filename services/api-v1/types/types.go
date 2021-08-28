package types

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrNotFound = errors.New("Not found")
)

type APIType struct {
	version string
	lang    string
}

type Profile struct {
	Hostname      string
	RemoteAddress string
	Author        string
	Hobbies       []string
}

type QueryCriteria struct {
	Criteria []map[string]string
}
type Visitor struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name,omitempty"`
	count int                `bson:"count,omitempty"`
}
