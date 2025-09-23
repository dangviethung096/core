package core

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

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

type CacheData struct {
	Key        string
	Value      any
	Expiration time.Duration
}

func SetCache(ctx context.Context, cacheData CacheData) error {
	var valueToStore any

	switch v := cacheData.Value.(type) {
	case string:
		valueToStore = v
	case int, int64, float64, bool:
		valueToStore = fmt.Sprintf("%v", v)
	default:
		// For structs, slices, maps, etc.
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			log.Printf("Error marshalling cache data: key %s, value %v, error %v", cacheData.Key, cacheData.Value, err)
			return err
		}
		valueToStore = string(jsonBytes)
	}

	err := redisClient.Client.Set(ctx, cacheData.Key, valueToStore, cacheData.Expiration).Err()
	if err != nil {
		log.Printf("Error setting cache: key %s, value %v, expiration %v, error %v", cacheData.Key, cacheData.Value, cacheData.Expiration, err)
		return err
	}

	log.Printf("Set cache: key %s, value %v, expiration %v", cacheData.Key, valueToStore, cacheData.Expiration)
	return nil
}

func GetCache[T any](ctx context.Context, key string) (T, error) {
	var result T
	strValue, err := redisClient.Client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Error getting cache: key %s, error %v", key, err)
		return result, err
	}

	// Use reflection to determine the type of T
	tType := reflect.TypeOf(result)
	switch tType.Kind() {
	case reflect.String:
		result = any(strValue).(T)
	case reflect.Bool:
		b, err := strconv.ParseBool(strValue)
		if err != nil {
			log.Printf("Error parsing bool cache data: key %s, value %s, error %v", key, strValue, err)
			return result, err
		}
		result = any(b).(T)
	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			log.Printf("Error parsing int cache data: key %s, value %s, error %v", key, strValue, err)
			return result, err
		}
		result = any(i).(T)
	case reflect.Float64:
		f, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			log.Printf("Error parsing float64 cache data: key %s, value %s, error %v", key, strValue, err)
			return result, err
		}
		result = any(f).(T)
	default:
		// Assume struct, try to unmarshal JSON
		err := json.Unmarshal([]byte(strValue), &result)
		if err != nil {
			log.Printf("Error parsing struct cache data: key %s, value %s, error %v", key, strValue, err)
			return result, err
		}
	}
	return result, nil
}

func DeleteCache(ctx context.Context, key string) error {
	err := redisClient.Client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Error deleting cache: key %s, error %v", key, err)
		return err
	}
	return nil
}

func DeleteCacheByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var err error

	for {
		var keys []string
		keys, cursor, err = redisClient.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("Error scanning cache keys with pattern %s: error %v", pattern, err)
			return err
		}

		if len(keys) > 0 {
			err = redisClient.Client.Del(ctx, keys...).Err()
			if err != nil {
				log.Printf("Error deleting cache keys with pattern %s: keys %v, error %v", pattern, keys, err)
				return err
			}
			log.Printf("Deleted %d cache keys with pattern %s", len(keys), pattern)
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
