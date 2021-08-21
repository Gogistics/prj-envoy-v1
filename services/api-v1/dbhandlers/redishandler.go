package dbhandlers

import (
    "fmt"
    "log"
    "context"
    "github.com/go-redis/redis/v8"
)


type RedisClientWrapper struct {
  Client *redis.Client
}


var (
  ctx = context.Background()
  RedisWrapper = RedisClientWrapper{
    redis.NewClient(&redis.Options{
        Addr:     "172.10.0.61:6379", // 172.10.0.61 is the IP of redis standalone and 172.10.0.50 is the IP of envoy proxy
        Password: "", // no password set
        DB:       0,  // use default DB
    })}
)

func (wrapper RedisClientWrapper) Set(key string, val []byte) error {
  err := wrapper.Client.Set(ctx, key, val, 0).Err()
  if err != nil {
    log.Fatal(err)
  }
  return err
}

func (wrapper RedisClientWrapper) Get(key string) ([]byte, error) {
  val, err := wrapper.Client.Get(ctx, key).Result()
  if err == redis.Nil {
    fmt.Println("key does not exist")
  } else if err != nil {
    log.Fatal(err)
  } else {
    fmt.Println("val: ", val)
  }

  return []byte(val), nil
}

