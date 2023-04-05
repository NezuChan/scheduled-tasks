package config

import (
	"github.com/caarlos0/env/v7"
	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/lib/redis"
)

type Config struct {
	AMQPUrl string `env:"AMQP_URL,required"`

	Redis redis.Config
}

func Init() (conf Config, err error) {
	conf = Config{}
	if err := env.Parse(&conf); err != nil {
		log.Fatalf("%+v\n", err)
	}
	return conf, nil
}
