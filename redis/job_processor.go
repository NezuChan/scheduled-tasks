package redis

import (
	"context"
	"fmt"
	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/broker"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/nezuchan/scheduled-tasks/snowflake"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func ProcessJob(client redis.UniversalClient, broker broker.Broker) {
	AddTask(client, snowflake.GenerateNew(), time.Now().Add(10*time.Second), "o")

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
			key, _ := fmt.Printf("scheduler_value:%s", taskId)
			value, err := client.Get(context.Background(), string(rune(key))).Result()

			if err != nil {
				panic(err)
			}
			log.Infof("Running task with ID %s %s", taskId, value)
			// TODO: Send the task to client. tell client that this task need to be executed on theirs.
			broker.Channel.PublishWithContext(context.Background(), constants.TASKER_EXCHANGE)
		}
	}()
}

func AddTask(client redis.UniversalClient, taskId string, runTime time.Time, data any) {
	err := client.ZAdd(context.Background(), "scheduler", redis.Z{
		Score:  float64(runTime.Unix()),
		Member: taskId,
	}).Err()

	key, _ := fmt.Printf("scheduler_value:%s", taskId)
	err = client.Set(context.Background(), string(rune(key)), data, 0).Err()

	if err != nil {
		panic(err)
	}
	log.Infof("Added task with ID %s to run at %v.\n", taskId, runTime)
}
