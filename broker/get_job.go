package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/redis/go-redis/v9"
)

func GetJob(client redis.UniversalClient,broker Broker, message Message) ([]byte, error) {
	if data, ok := message.D.Data.(map[string]interface{}); ok {
    	if taskId, ok := data["taskId"].(string); ok {

			value, err := client.Get(context.Background(), fmt.Sprintf("%s:%s", constants.TASK_REDIS_KEY_VALUE, taskId)).Result()

			if err == redis.Nil {
				return json.Marshal(map[string]interface{}{
					"t": constants.TASK_GET,
					"d": map[string]interface{}{
						"taskId": taskId,
						"message": "task not found",
					},
				})
			}

			log.Infof("Requested task get with ID %s", taskId)

			var msg any

			err = json.Unmarshal([]byte(value), &msg)

			if err != nil {
				panic(err)
			}

			return json.Marshal(map[string]interface{}{
				"t": constants.TASK_GET,
				"d": msg,
			})
    	} else {
        	return json.Marshal(map[string]interface{}{
				"t": constants.TASK_GET,
				"d": map[string]interface{}{
					"message": "taskId is not a string or does not exist",
				},
			})
    	}
	} else {
    	return json.Marshal(map[string]interface{}{
			"t": constants.TASK_GET,
			"d": map[string]interface{}{
				"message": "taskId is not a string or does not exist",
			},
		})
	}
}