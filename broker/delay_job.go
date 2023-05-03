package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/nezuchan/scheduled-tasks/snowflake"
	"github.com/redis/go-redis/v9"
)

func DelayJob(client redis.UniversalClient,broker Broker, message Message) ([]byte, error) {
	later := time.Now().Add(time.Duration((message.D.Time / 1000) * int(time.Second)))
	taskId := snowflake.GenerateNew()
	
	err := client.ZAdd(context.Background(), constants.TASK_REDIS_KEY, redis.Z{
		Score:  float64(later.Unix()),
		Member: taskId,
	}).Err()
	
	if err != nil {
		panic(err)
	}

	value, err := json.Marshal(message.D.Data)

	if err != nil {
		panic(err)
	}

	err = client.Set(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_VALUE, taskId), value, 0).Err()

	if err != nil {
		panic(err)
	}

	if message.D.Route != nil {
		err = client.Set(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId), *message.D.Route, 0).Err()

		if err != nil {
			panic(err)
		}
	}

	log.Infof("Added task with ID %s to run at %v.\n", taskId, later)

	return json.Marshal(map[string]interface{}{
		"t": constants.TASK_DELAY,
		"d": map[string]interface{}{
			"taskId": taskId,
		},
	})
}