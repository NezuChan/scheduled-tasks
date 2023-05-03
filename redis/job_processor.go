package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/broker"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/nezuchan/scheduled-tasks/snowflake"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func ProcessJob(client redis.UniversalClient, broker broker.Broker) {
	AddTask(client, snowflake.GenerateNew(), time.Now().Add(10*time.Second), "{\"t\":\"delay\",\"d\":{\"op\":2}}")

	go func() {
		for {
			now := time.Now()
			result, err := client.ZRangeByScore(context.Background(), "scheduler", &redis.ZRangeBy{
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
			err = client.ZRem(context.Background(), "scheduler", taskId).Err()

			if err != nil {
				panic(err)
			}

			key := fmt.Sprintf("scheduler_value:%s", taskId)
			value, err := client.Get(context.Background(), key).Result()
			client.Unlink(context.Background(), key)

			if err != nil {
				panic(err)
			}

			log.Infof("Sending task to client with taskId %s", taskId)
			// TODO: Add custom router from client data sent from. and probably docker replica support?
			broker.Channel.PublishWithContext(context.Background(), constants.TASKER_EXCHANGE, "*", false, false, amqp091.Publishing{
				ContentType: "text/plain",
				Body:        []byte(value),
			})
		}
	}()
}

func AddTask(client redis.UniversalClient, taskId string, runTime time.Time, data any) {
	err := client.ZAdd(context.Background(), "scheduler", redis.Z{
		Score:  float64(runTime.Unix()),
		Member: taskId,
	}).Err()
	
	if err != nil {
		panic(err)
	}

	err = client.Set(context.Background(), fmt.Sprintf("scheduler_value:%s", taskId), data, 0).Err()

	if err != nil {
		panic(err)
	}

	log.Infof("Added task with ID %s to run at %v.\n", taskId, runTime)
}
