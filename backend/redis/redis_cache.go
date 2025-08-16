package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

var (
	redisOnce     sync.Once
	redisInstance *RedisCache
)

func GetRedisCache() *RedisCache {
	redisOnce.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr: "redis:6379",
			DB:   0,
		})
		redisInstance = &RedisCache{client: client}
	})
	return redisInstance
}

func (r *RedisCache) HashCode(code string) string {
	h := sha256.New()
	h.Write([]byte(code))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *RedisCache) CacheCode(hash, code string) error {
	ctx := context.Background()
	return r.client.Set(ctx, "code:"+hash, code, 24*time.Hour).Err()
}

func (r *RedisCache) CacheResult(taskID string, result interface{}) error {
	ctx := context.Background()
	data, _ := json.Marshal(result)
	return r.client.Set(ctx, "result:"+taskID, data, 60*time.Second).Err()
}

func (r *RedisCache) GetCachedResult(taskID string, dest interface{}) error {
	ctx := context.Background()
	val, err := r.client.Get(ctx, "result:"+taskID).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}
