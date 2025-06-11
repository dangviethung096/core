package core

import (
	"crypto/tls"
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
	redisOptions := &redis.Options{
		Addr:     address,
		Password: BLANK,
		DB:       0,
	}

	if Config.Redis.SecureConnection {
		redisOptions.TLSConfig = &tls.Config{}
	}

	if Config.Redis.Password != BLANK {
		redisOptions.Password = Config.Redis.Password
	}

	client := redis.NewClient(redisOptions)
	return cacheClient{
		Client: client,
	}
}
