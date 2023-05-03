package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/broker"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func ProcessJob(client redis.UniversalClient, broker broker.Broker) {
	go func() {
		for {
			now := time.Now()
			result, err := client.ZRangeByScore(context.Background(), constants.TASK_REDIS_KEY, &redis.ZRangeBy{
				Min: "0",
				Max: strconv.FormatInt(now.Unix(), 10),
			}).Result()
			if err != nil {
				panic(err)
			}
			if len(result) == 0 {
				time.Sleep(time.Second / 100)
				continue
			}
			taskId := result[0]
			err = client.ZRem(context.Background(), constants.TASK_REDIS_KEY, taskId).Err()

			if err != nil {
				panic(err)
			}

			key := fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_VALUE, taskId)
			value, err := client.Get(context.Background(), key).Result()
			client.Unlink(context.Background(), key)

			if err != nil {
				log.Warnf("Failed to get task with taskId %s", taskId)
				return
			}

			route, err := client.Get(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId)).Result()
			client.Unlink(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId))
			log.Infof("Sending task to client with taskId %s", taskId)

			// TODO: Add docker replica support?

			if err != redis.Nil {
				broker.Channel.PublishWithContext(context.Background(), constants.TASKER_EXCHANGE, route, false, false, amqp091.Publishing{
					ContentType: "text/plain",
					Body:        []byte(value),
				})
			} else {
				broker.Channel.PublishWithContext(context.Background(), constants.TASKER_EXCHANGE, "*", false, false, amqp091.Publishing{
					ContentType: "text/plain",
					Body:        []byte(value),
				})
			}
		}
	}()
}
