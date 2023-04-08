package redis

import (
	"context"
	"github.com/disgoorg/log"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func ProcessJob(client redis.UniversalClient) {
	addTask(client, "task-3", time.Now().Add(3*time.Second))

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
			taskID := result[0]
			err = client.ZRem(context.Background(), "scheduler", taskID).Err()
			if err != nil {
				panic(err)
			}
			log.Infof("Running task with ID %s", taskID)
			// TODO: Run the task here
		}
	}()
}

func addTask(client redis.UniversalClient, taskID string, runTime time.Time) {
	err := client.ZAdd(context.Background(), "scheduler", redis.Z{
		Score:  float64(runTime.Unix()),
		Member: taskID,
	}).Err()
	if err != nil {
		panic(err)
	}
	log.Infof("Added task with ID %s to run at %v.\n", taskID, runTime)
}
