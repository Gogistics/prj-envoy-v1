package dbhandlers

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClientWrapper struct {
	Client *redis.Client
}

/* Ref. for using Redis:
   https://github.com/go-redis/redis
   https://redis.com/blog/connection-pools-for-serverless-functions-and-backend-services/
   https://github.com/go-redis/redis/blob/master/example_test.go
*/
var (
	ctx = context.Background()
	// 172.10.0.61 is the IP of redis standalone and 172.10.0.50 is the IP of envoy proxy
	RedisWrapper = RedisClientWrapper{
		redis.NewClient(&redis.Options{
			Addr:         "172.10.0.50:6379",
			Password:     "", // no password set
			DB:           0,  // use default DB
			DialTimeout:  10 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PoolSize:     20,
			PoolTimeout:  30 * time.Second,
		})}
)

func (wrapper RedisClientWrapper) Set(key string, val []byte) error {
	err := wrapper.Client.Set(ctx, key, val, 0).Err()
	if err != nil {
		log.Fatalln("Error: failed to set key-value pair. ", err)
	}
	return err
}

func (wrapper RedisClientWrapper) Incr(key string) (int64, error) {
	count, err := wrapper.Client.Incr(ctx, key).Result()
	if err != nil {
		log.Fatalln("Error: failed to Incr key. ", err)
		return -1, err
	}
	return count, nil
}

func (wrapper RedisClientWrapper) Expire(key string, expiration time.Duration) error {
	_, err := wrapper.Client.Expire(ctx, key, expiration).Result()
	if err != nil {
		log.Fatalln("Error: failed to set key expiration. ", err)
		return err
	}
	return nil
}

func (wrapper RedisClientWrapper) Get(key string) ([]byte, error) {
	val, err := wrapper.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Println("Error: key does not exist")
	} else if err != nil {
		log.Fatal("Error: failed to query value by the given key. ", err)
	} else {
		log.Println("val: ", val)
	}

	return []byte(val), nil
}
