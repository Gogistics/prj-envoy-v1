package dbhandlers

import (
    "fmt"
    "log"
    "context"
    "time"
    "github.com/go-redis/redis/v8"
)


type RedisClientWrapper struct {
  Client *redis.Client
}


/* Ref:
    https://github.com/go-redis/redis
*/
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
    fmt.Println("Error: failed to set key-value pair")
    log.Fatal(err)
  }
  return err
}

func (wrapper RedisClientWrapper) Incr(key string) (int64, error) {
  count, err := wrapper.Client.Incr(ctx, key).Result()
  if err != nil {
    fmt.Println("Error: failed to Incr key")
    log.Fatal(err)
    return -1, err
  }
  return count, nil
}

func (wrapper RedisClientWrapper) Expire(key string, expiration time.Duration) error {
  _, err := wrapper.Client.Expire(ctx, key, expiration).Result()
  if err != nil {
    fmt.Println("Error: failed to set key expiration")
    return err
  }
  return nil
}

func (wrapper RedisClientWrapper) Get(key string) ([]byte, error) {
  val, err := wrapper.Client.Get(ctx, key).Result()
  if err == redis.Nil {
    fmt.Println("Error: key does not exist")
  } else if err != nil {
    fmt.Println("Error: failed to query value by the given key")
    log.Fatal(err)
  } else {
    fmt.Println("val: ", val)
  }

  return []byte(val), nil
}

