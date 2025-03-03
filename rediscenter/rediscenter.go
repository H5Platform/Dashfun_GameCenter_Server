package rediscenter

import (
	"context"
	"dashfun_gamecenter/config"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
)

var onceRedisCenter sync.Once
var instRedisCenter *RedisCenter

type RedisCenter struct {
	client *redis.Client
}

func Get() *redis.Client {
	onceRedisCenter.Do(func() {
		instRedisCenter = &RedisCenter{}
		instRedisCenter.init()
	})
	return instRedisCenter.client
}

func (r *RedisCenter) init() {
	cfg := config.GetConfig().RedisCfg
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("redis center init fail, err:%s", err.Error())
	}
	log.Println("redis center init success")
	r.client = rdb
}
