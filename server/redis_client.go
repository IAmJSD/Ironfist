package main

import (
	"github.com/go-redis/redis/v7"
	"os"
)

// RedisClient defines the Redis client.
var RedisClient = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_ADDR"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
})

// Runs a sanity check.
func init() {
	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}
