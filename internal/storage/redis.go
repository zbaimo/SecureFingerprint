package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

type UserScore struct {
	Score     int       `json:"score"`
	LastSeen  time.Time `json:"last_seen"`
	RequestCount int    `json:"request_count"`
}

type AccessLog struct {
	Fingerprint string    `json:"fingerprint"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Path        string    `json:"path"`
	Timestamp   time.Time `json:"timestamp"`
	Score       int       `json:"score"`
}

func NewRedisClient(addr, password string, db int, poolSize int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis连接失败: %v", err)
	}

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// 获取用户分数
func (r *RedisClient) GetUserScore(fingerprint string) (*UserScore, error) {
	key := fmt.Sprintf("user_score:%s", fingerprint)
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		// 用户不存在，返回默认分数
		return &UserScore{
			Score:        100,
			LastSeen:     time.Now(),
			RequestCount: 0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var score UserScore
	err = json.Unmarshal([]byte(val), &score)
	return &score, err
}

// 更新用户分数
func (r *RedisClient) UpdateUserScore(fingerprint string, score *UserScore) error {
	key := fmt.Sprintf("user_score:%s", fingerprint)
	data, err := json.Marshal(score)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, 24*time.Hour).Err()
}

// 记录访问日志到Redis（短期存储）
func (r *RedisClient) LogAccess(log *AccessLog) error {
	key := fmt.Sprintf("access_log:%s:%d", log.Fingerprint, log.Timestamp.Unix())
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, time.Hour).Err()
}

// 获取用户最近访问记录
func (r *RedisClient) GetRecentAccess(fingerprint string, minutes int) ([]AccessLog, error) {
	pattern := fmt.Sprintf("access_log:%s:*", fingerprint)
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var logs []AccessLog
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)

	for _, key := range keys {
		val, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			continue
		}

		var log AccessLog
		if err := json.Unmarshal([]byte(val), &log); err != nil {
			continue
		}

		if log.Timestamp.After(cutoff) {
			logs = append(logs, log)
		}
	}

	return logs, nil
}

// 检查用户是否被封禁
func (r *RedisClient) IsUserBanned(fingerprint string) (bool, time.Duration, error) {
	key := fmt.Sprintf("banned:%s", fingerprint)
	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		return false, 0, err
	}
	
	if ttl <= 0 {
		return false, 0, nil
	}
	
	return true, ttl, nil
}

// 封禁用户
func (r *RedisClient) BanUser(fingerprint string, duration time.Duration) error {
	key := fmt.Sprintf("banned:%s", fingerprint)
	return r.client.Set(r.ctx, key, "banned", duration).Err()
}

// 解除封禁
func (r *RedisClient) UnbanUser(fingerprint string) error {
	key := fmt.Sprintf("banned:%s", fingerprint)
	return r.client.Del(r.ctx, key).Err()
}

// 获取访问频率（每分钟请求数）
func (r *RedisClient) GetRequestRate(fingerprint string) (int, error) {
	key := fmt.Sprintf("rate:%s", fingerprint)
	count, err := r.client.Get(r.ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// 增加请求计数
func (r *RedisClient) IncrementRequestRate(fingerprint string) error {
	key := fmt.Sprintf("rate:%s", fingerprint)
	pipe := r.client.Pipeline()
	pipe.Incr(r.ctx, key)
	pipe.Expire(r.ctx, key, time.Minute)
	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
