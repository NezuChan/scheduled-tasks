package lib

import (
	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/broker"
	"github.com/nezuchan/scheduled-tasks/config"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/nezuchan/scheduled-tasks/redis"
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

	err := task.Broker.Channel.ExchangeDeclare(
		constants.TASKER_EXCHANGE,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Unable to declare exchange due to: %v", err)
	}

	redis.ProcessJob(task.Redis, *task.Broker); broker.HandleReceive(*task.Broker)
	return &task
}
