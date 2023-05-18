package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/constants"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Broker struct {
	Connection   *amqp091.Connection
	Channel      *amqp091.Channel
	retryAttempt int
}

func CreateBroker(amqpURI string) *Broker {
	broker := &Broker{}
	var err error

	for {
		broker.Connection, err = amqp091.DialConfig(amqpURI,
			amqp091.Config{
				Dial: amqp091.DefaultDial(time.Second * 15),
			})

		if err == nil {
			break
		}

		log.Fatalf("Failed to connect to RabbitMQ: %v, retrying in %d seconds", err, 10)
		time.Sleep(time.Second * 10)
	}

	broker.Channel, err = broker.Connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}

	go broker.handleReconnect(amqpURI)
	log.Infof("Connected to RabbitMQ server")
	return broker
}

func (broker *Broker) handleReconnect(amqpURI string) {
	onClose := broker.Connection.NotifyClose(make(chan *amqp091.Error))
	for {
		<-onClose
		log.Warnf("RabbitMQ connection lost. Reconnecting...")
		for {
			if broker.retryAttempt >= 3 {
				panic(fmt.Sprintf("couldn't reconnect to amqp broker after %d attempts", broker.retryAttempt))
			}
			conn, err := amqp091.DialConfig(amqpURI, amqp091.Config{
				Dial: amqp091.DefaultDial(time.Second * 5),
			})
			broker.retryAttempt += 1
			if err == nil {
				broker.Connection = conn
				broker.retryAttempt = 0
				break
			}
			log.Fatalf("Failed to connect to RabbitMQ: %v, retrying in %d seconds", err, 5)
			time.Sleep(time.Second * 5)
		}

		log.Infof("Re-connected to RabbitMQ server")
		broker.Channel, _ = broker.Connection.Channel()
		go broker.handleReconnect(amqpURI)
	}
}


func HandleReceive(redis redis.UniversalClient, broker Broker) {
	q, err := broker.Channel.QueueDeclare(
		constants.TASKER_SEND,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Unable to declare queue due to: %v", err)
	}

	err = broker.Channel.Qos(
		1,
		0,
		false,
	)

	if err != nil {
		log.Fatalf("Failed to set QoS due to: %v", err)
	}

	messages, err := broker.Channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to consume message due to: %v", err)
	}

	for d := range messages {
		go func(delivery amqp091.Delivery) {

			var m Message
			err := json.Unmarshal([]byte(delivery.Body), &m)

			if err != nil {
				log.Fatalf("Failed to unmarshal string: %v", err)
			}

			log.Debugf("Received message: %s", delivery.Body)

			switch m.T {
				case constants.TASK_DELAY: {
					log.Debugf("Received a action to create delated task: %s", m.D)

					value, err := DelayJob(redis, broker, m)

					if err != nil {
						log.Fatalf("Failed to create delayed task due to: %v", err)
					}

					go ReplyBack(broker, delivery, value)
					break
				}

				case constants.TASK_DELETE: {
					log.Debugf("Received a action to delete task: %s", m.D)

					value, err := DeleteJob(redis, broker, m)

					if err != nil {
						log.Fatalf("Failed to delete task due to: %v", err)
					}

					go ReplyBack(broker, delivery, value)
					break
				}

				case constants.TASK_GET: {
					log.Debugf("Received a action to get task: %s", m.D)

					value, err := GetJob(redis, broker, m)

					if err != nil {
						log.Fatalf("Failed to get task due to: %v", err)
					}

					go ReplyBack(broker, delivery, value)
					break
				}

				case constants.TASK_CRON: {
					log.Debugf("Received a action to get task: %s", m.D)

					value, err := CronJob(redis, broker, m)

					if err != nil {
						log.Fatalf("Failed to create cron job task due to: %v", err)
					}

					go ReplyBack(broker, delivery, value)
					break
				}

				default: {
					log.Warnf("Received a unknown action: %s", m.T)
					
					value, err := json.Marshal(map[string]interface{}{
						"message": "unknown action",
					})

					if err != nil {
						log.Fatalf("Failed to marshal string: %v", err)
					}

					go ReplyBack(broker, delivery, value)

					break
				}
			}
		}(d)
	}
}

func ReplyBack(broker Broker, delivery amqp091.Delivery, value []byte) {
	err := broker.Channel.PublishWithContext(context.Background(),
		"",
		delivery.ReplyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			CorrelationId: delivery.CorrelationId,
			Body: value,
		},
	)

	if err != nil {
		log.Fatalf("Failed to publish message due to: %v", err)
	}

	delivery.Ack(false)
}

type Message struct {
	T string `json:"t"`
	D struct {
		Time  interface{} `json:"time"`
		Data  interface{} `json:"data"`
		Name  *string     `json:"name"`
		Route *string     `json:"route"`
	} `json:"d"`
}