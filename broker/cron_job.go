package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/nezuchan/scheduled-tasks/processor"
	"github.com/nezuchan/scheduled-tasks/snowflake"
	"github.com/redis/go-redis/v9"
)

func CronJob(client redis.UniversalClient, broker Broker, message Message) ([]byte, error) {
	var timeVal string
	switch t := message.D.Time.(type) {
	case string:
		timeVal = t
	default:
		return json.Marshal(map[string]interface{}{
			"t": constants.TASK_CRON,
			"d": map[string]interface{}{
				"message": "Not a string",
			},
		})
	}

	if message.D.Name == nil {
		return json.Marshal(map[string]interface{}{
			"t": constants.TASK_CRON,
			"d": map[string]interface{}{
				"message": "Name is required",
			},
		})
	}

	taskId := snowflake.GenerateNew()
	value, err := json.Marshal(message.D.Data)

	if err != nil {
		panic(err)
	}

	isExist := client.Exists(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_VALUE, *message.D.Name)).Val()

	if isExist == 1 {
		return json.Marshal(map[string]interface{}{
			"t": constants.TASK_CRON,
			"d": map[string]interface{}{
				"message": "That cron job already exists",
			},
		})
	}

	client.Set(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId), value, 0).Err()
	client.Set(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_VALUE, *message.D.Name), value, 0).Err()

	if message.D.Route != nil {
		err = client.Set(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId), *message.D.Route, 0).Err()

		if err != nil {
			panic(err)
		}
	}

	client.SAdd(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_CRON_SETS, *message.D.Name), taskId).Err()

	log.Infof("Added task with ID %s to run every %v.\n", taskId, timeVal)

	processor.ProcessCronJob(client, *broker.Channel, *message.D.Name, taskId)

	return json.Marshal(map[string]interface{}{
		"t": constants.TASK_DELAY,
		"d": map[string]interface{}{
			"taskId": taskId,
		},
	})
}