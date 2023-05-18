package lib

import (
	"context"
	"strings"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/broker"
	"github.com/nezuchan/scheduled-tasks/config"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/nezuchan/scheduled-tasks/processor"
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
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Unable to declare exchange due to: %v", err)
	}

	processor.ProcessJob(task.Redis, *task.Broker.Channel); broker.HandleReceive(task.Redis, *task.Broker)

	Members := task.Redis.SMembers(context.Background(), constants.TASK_REDIS_CRON_SETS).Val()

	for _, member := range Members {
		taskId := strings.Split(member, ":")[0]
		name := strings.Split(member, ":")[1]

		processor.ProcessCronJob(task.Redis, *task.Broker.Channel, name, taskId)
	}
	return &task
}
