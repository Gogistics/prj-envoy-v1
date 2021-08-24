package dbhandlers

import (
  "context"
  "fmt"
  "log"
  "time"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClientWrapper struct {
  Client *mongo.Client
}

var (
  MongoWrapper = MongoClientWrapper{
    getMongoClient()}
)

func getMongoClient() *mongo.Client {
  // 172.10.0.71 is the IP of mongo standalone and 172.10.0.55 is the IP of envoy proxy
  mongoURI := "mongodb://web_test_user:web-1234567@172.10.0.55:27017/web"
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
  if err != nil {
    panic(err)
  }
  return client
}

func (wrapper MongoClientWrapper) InsertOne(val string) error {
  coll := wrapper.Client.Database("web").Collection("test")
  resp, err := coll.InsertOne(context.TODO(), bson.M{"userName": val})
  if err != nil {
    fmt.Println("Error: failed to insert data into document")
    return err
  }
  fmt.Println("InsertedID: ", resp.InsertedID)
  return nil
}

/* Ref:
    https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#Collection.FindOneAndUpdate
*/
func (wrapper MongoClientWrapper) FindOneAndUpdate(key string, val string, newData *map[string]string ) (string, error) {
  coll := wrapper.Client.Database("web").Collection("test")
  opts := options.FindOneAndUpdate().SetUpsert(true)
  filter := bson.D{{key, val}}

  var setNewData bson.D
  if newData != nil {
    for key, val := range *newData {
      setNewData = append(setNewData, bson.E{key, val})
    }
  }
  var update bson.D
  update = bson.D{{"$inc", bson.D{{"count", 1}}}, {"$set", setNewData}}
  var updatedDocument bson.M
  err := coll.FindOneAndUpdate(
    context.TODO(),
    filter,
    update, opts,
  ).Decode(&updatedDocument)

  if err != nil {
    log.Fatal(err)
    if err == mongo.ErrNoDocuments {
      return "Error: no document in the collection", err
    }
    return "Error: failed to find document", err
  }

  id, ok := updatedDocument["_id"]
  if !ok {
    fmt.Println("_id does not exist in the updatedDocument map")
  }
  strObjId := id.(primitive.ObjectID).Hex()
  return strObjId, nil
}

/* Ref:
    https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#Collection.Find
*/
func (wrapper MongoClientWrapper) Find(key *string, val *string) ([]bson.M) {
  coll := wrapper.Client.Database("web").Collection("test")
  opts := options.Find()
  if key != nil {
    opts = options.Find().SetSort(bson.D{{*key, 1}})
  }
  filter := bson.D{}
  if key != nil || val != nil {
    filter = append(filter, bson.E{*key, *val})
  }
  cursor, err := coll.Find(context.TODO(), filter, opts)

  if err != nil {
    log.Fatal(err)
    fmt.Println("Error: failed to query data")
  }

  results := []bson.M{}
  if err = cursor.All(context.TODO(), &results); err != nil {
    log.Fatal(err)
    fmt.Println("Error: failed to construct returned data")
  }
  return results
}

func (wrapper MongoClientWrapper) FindID(key string, val string) (string, error) {
  var result bson.M

  coll := wrapper.Client.Database("web").Collection("test")
  err := coll.FindOne(context.TODO(), bson.M{key: val}).Decode(&result)
  if err != nil {
    if err == mongo.ErrNoDocuments {
      return "Error: no document in the collection", err
    }
    return "Error: failed to find document", err
  }

  id, ok := result["_id"]
  if !ok {
    fmt.Println("_id does not exist in result map")
  }
  strObjID := id.(primitive.ObjectID).Hex()

  return strObjID, nil
}


