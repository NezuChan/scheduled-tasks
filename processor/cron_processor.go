package processor

import (
	"context"
	"fmt"

	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
)

func ProcessCronJob(client redis.UniversalClient, broker amqp091.Channel, name string, taskId string) {
	CronValue := client.Exists(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_VALUE, name)).Val()
	TaskKey   := client.Exists(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId)).Val()

	if CronValue == 1 && TaskKey == 1 {
		Cron := client.Get(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_VALUE, name)).Val()

		c := cron.New()

		c.AddFunc(Cron, func() {
			Value := client.Get(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId)).Val()
			
			if Value == "" {
				c.Stop()
				client.Unlink(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId))
				client.Unlink(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_VALUE, name))
				client.SRem(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_SETS, name)).Err()
				return
			}

			Route := client.Get(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId)).Val()

			if Route != "" {
				broker.PublishWithContext(context.Background(), constants.TASKER_EXCHANGE, Route, false, false, amqp091.Publishing{
					ContentType: "text/plain",
					Body:        []byte(Value),
				})
			} else {
				broker.PublishWithContext(context.Background(), constants.TASKER_EXCHANGE, "*", false, false, amqp091.Publishing{
					ContentType: "text/plain",
					Body:        []byte(Value),
				})
			}
		})

		c.Start()

	} else {
		client.Unlink(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId))
		client.Unlink(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_VALUE, name))
		client.SRem(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_SETS, name)).Err()
	}
}