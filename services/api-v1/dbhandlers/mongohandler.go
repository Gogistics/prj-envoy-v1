package dbhandlers

import (
  "context"
  "fmt"
  "log"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCLientWrapper struct {
  Client *mongo.Client
}

var (
  MongoWrapper = MongoCLientWrapper{
    mongo.Connect(context.TODO(), options.Client()
      .ApplyURI("mongodb://dev.mongo.com:27017")
    )}
)