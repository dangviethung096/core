package core

import (
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

type cacheClient struct {
	*redis.Client
}

func connectCacheDB() cacheClient {
	address := fmt.Sprintf("%s:%d", Config.Redis.Host, Config.Redis.Port)
	log.Printf("Connecting to redis at %s\n", address)
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: BLANK,
		DB:       0,
	})
	return cacheClient{
		Client: client,
	}
}
