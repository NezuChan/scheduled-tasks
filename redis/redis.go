package redis

import (
	"context"
	"fmt"
	"github.com/disgoorg/log"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type URLs []string

func (r *URLs) UnmarshalText(text []byte) error {
	str := strings.TrimPrefix(string(text), "[")
	str = strings.TrimSuffix(str, "]")
	str = strings.ReplaceAll(str, string('"'), "")
	str = strings.ReplaceAll(str, "redis://", "")
	*r = strings.Split(str, ", ")
	return nil
}

type Config struct {
	Password *string `env:"REDIS_PASSWORD" envDefault:""`
	Username *string `env:"REDIS_USERNAME" envDefault:""`
	PoolSize *int    `env:"REDIS_POOL_SIZE"`
	Db       *int    `env:"REDIS_DB" envDefault:"0"`
	Clusters *URLs   `env:"REDIS_CLUSTERS,required"`
}

type Redis struct {
	redis.UniversalClient
	config Config
}

func InitRedis(conf Config) *Redis {
	opts := &redis.UniversalOptions{
		Addrs: *conf.Clusters,
		DB:    *conf.Db,
	}

	if conf.Username != nil {
		opts.Username = *conf.Username
	}

	if conf.Password != nil {
		opts.Password = *conf.Password
	}

	if conf.PoolSize != nil {
		opts.PoolSize = *conf.PoolSize
	}

	client := redis.NewUniversalClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Info(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to Redis: %s", err))
	}

	log.Infof("Connected to Redis server")

	return &Redis{UniversalClient: client, config: conf}
}
