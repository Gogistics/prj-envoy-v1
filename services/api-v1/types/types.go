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

/* ref:
https://www.mongodb.com/blog/post/quick-start-golang--mongodb--modeling-documents-with-go-data-structures
https://youtu.be/leNCfU5SYR8
*/

type QueryCriteria struct {
	Criteria []map[string]string
}

/*
Notes: Visitor is an example for Mongo schema validation
- annotation: `bson:"..."`
- bson (binary JSON): a glue between Go data structure and JSON

Schema design of Mongo:
- techniques:
	- embedding
	- referencing
- considerations:
	- how to store data
	- query performance

Ref:
- https://youtu.be/leNCfU5SYR8
- https://www.mongodb.com/json-and-bson
- https://flaviocopes.com/go-tags/
- https://stackoverflow.com/questions/10858787/what-are-the-uses-for-tags-in-go/30889373#30889373
*/
type Visitor struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name,omitempty"`
	Count int                `bson:"count,omitempty"`
}
