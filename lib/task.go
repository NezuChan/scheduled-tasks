package lib

import (
	"github.com/nezuchan/scheduled-tasks/config"
	"github.com/nezuchan/scheduled-tasks/lib/broker"
	"github.com/nezuchan/scheduled-tasks/lib/redis"
)

type Task struct {
	Broker *broker.Broker
	Redis  *redis.Redis
}

func InitTask(conf *config.Config) *Task {
	task := Task{
		Broker: broker.CreateBroker(conf.AMQPUrl),
		Redis:  redis.InitRedis(conf.Redis),
	}

	return &task
}
