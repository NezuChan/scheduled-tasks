package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/redis/go-redis/v9"
)

func DeleteJob(client redis.UniversalClient,broker Broker, message Message) ([]byte, error) {
	if data, ok := message.D.Data.(map[string]interface{}); ok {
    	if taskId, ok := data["taskId"].(string); ok {

			value := client.Get(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_VALUE, taskId)).Val()

			if value == "" {
				return json.Marshal(map[string]interface{}{
					"t": constants.TASK_DELETE,
					"d": map[string]interface{}{
						"taskId": taskId,
						"message": "task not found",
					},
				})
			}

			err := client.ZRem(context.Background(), constants.TASK_REDIS_KEY, taskId).Err()

			if err != nil {
				panic(err)
			}

			err = client.Del(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_VALUE, taskId)).Err()

			if err != nil {
				panic(err)
			}

			err = client.Del(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_ROUTE, taskId)).Err()

			if err != nil {
				panic(err)
			}

			log.Infof("Deleted task with ID %s.\n", taskId)

			return json.Marshal(map[string]interface{}{
				"t": constants.TASK_DELETE,
				"d": map[string]interface{}{
					"taskId": taskId,
					"message": "deleted the task",
				},
			})
    	} else {
        	return json.Marshal(map[string]interface{}{
				"t": constants.TASK_DELETE,
				"d": map[string]interface{}{
					"message": "taskId is not a string or does not exist",
				},
			})
    	}
	} else {
    	return json.Marshal(map[string]interface{}{
			"t": constants.TASK_DELETE,
			"d": map[string]interface{}{
				"message": "taskId is not a string or does not exist",
			},
		})
	}
}