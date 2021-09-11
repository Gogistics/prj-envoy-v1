package dbhandlers

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

/* A Redis handler for handling data caching

Ref. for using Redis:
- https://github.com/go-redis/redis
- https://redis.com/blog/connection-pools-for-serverless-functions-and-backend-services/
- https://github.com/go-redis/redis/blob/master/example_test.go
*/

// RedisClientWrapper a struct of redis client wrapper
type RedisClientWrapper struct {
	Client *redis.Client
}

var (
	ctx = context.Background()
	// RedisWrapper a object of redis client wrapper
	// 172.10.0.61 is the IP of redis standalone and 172.10.0.50 is the IP of envoy proxy
	RedisWrapper = initRedisClientWrapper()
)

func initRedisClientWrapper() RedisClientWrapper {
	return RedisClientWrapper{
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
}

// Set simple key-value pair
func (wrapper *RedisClientWrapper) Set(key string, val []byte) error {
	err := wrapper.Client.Set(ctx, key, val, 0).Err()
	if err != nil {
		log.Fatalln("Error: failed to set key-value pair. ", err)
	}
	return err
}

// Incr increase count of the key by one
func (wrapper *RedisClientWrapper) Incr(key string) (int64, error) {
	count, err := wrapper.Client.Incr(ctx, key).Result()
	if err != nil {
		log.Fatalln("Error: failed to Incr key. ", err)
		return -1, err
	}
	return count, nil
}

// Expire set expiration of key
func (wrapper *RedisClientWrapper) Expire(key string, expiration time.Duration) error {
	_, err := wrapper.Client.Expire(ctx, key, expiration).Result()
	if err != nil {
		log.Fatalln("Error: failed to set key expiration. ", err)
		return err
	}
	return nil
}

// Get retrieve value by the key
func (wrapper *RedisClientWrapper) Get(key string) ([]byte, error) {
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
