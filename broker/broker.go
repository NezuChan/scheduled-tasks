package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/log"
	"github.com/nezuchan/scheduled-tasks/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	retryAttempt int
}

func CreateBroker(amqpURI string) *Broker {
	broker := &Broker{}
	var err error

	for {
		broker.Connection, err = amqp.DialConfig(amqpURI,
			amqp.Config{
				Dial: amqp.DefaultDial(time.Second * 15),
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
	onClose := broker.Connection.NotifyClose(make(chan *amqp.Error))
	for {
		<-onClose
		log.Warnf("RabbitMQ connection lost. Reconnecting...")
		for {
			if broker.retryAttempt >= 3 {
				panic(fmt.Sprintf("couldn't reconnect to amqp broker after %d attempts", broker.retryAttempt))
			}
			conn, err := amqp.DialConfig(amqpURI, amqp.Config{
				Dial: amqp.DefaultDial(time.Second * 5),
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


func HandleReceive(broker Broker) {
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
		go func(delivery amqp.Delivery) {
			log.Infof("Received message: %s", delivery.Body)
			err = broker.Channel.PublishWithContext(context.Background(),
				"",
				delivery.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					CorrelationId: delivery.CorrelationId,
					Body:        []byte("Hello World!"),
				},
			)

			if err != nil {
				log.Fatalf("Failed to publish message due to: %v", err)
			}

			delivery.Ack(false)
		}(d)
	}
}